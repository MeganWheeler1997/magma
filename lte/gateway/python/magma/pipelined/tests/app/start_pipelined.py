"""
Copyright (c) 2018-present, Facebook, Inc.
All rights reserved.

This source code is licensed under the BSD-style license found in the
LICENSE file in the root directory of this source tree. An additional grant
of patent rights can be found in the PATENTS file in the same directory.
"""

import logging
import threading
import subprocess
from enum import Enum
from collections import namedtuple
from concurrent.futures import Future

from magma.pipelined.rule_mappers import RuleIDToNumMapper
from magma.pipelined.app.base import MagmaController, ControllerType
from magma.pipelined.tests.app.exceptions import ServiceRunningError,\
    BadConfigError

from ryu.base.app_manager import AppManager
from ryu.lib import hub


class TestSetup(object):
    """
    The TestSetup class variables
        apps:               [Controller]:   ryu apps to instantiate
        references:         [Controller]:   futures to get references of
                                            instantiated apps
        config:             dict:           config for ryu app
        mconfig:            dict:           mconfig for ryu app
        service_manager:    ServiceManager: service manager for ryu app
        integ_test:         bool:           set true when running tests in
                                            integ setting
    """
    def __init__(self, apps, references, config, mconfig, loop,
                 service_manager, integ_test=False, rpc_stubs=None):
        self.apps = apps
        self.references = references
        self.config = config
        self.mconfig = mconfig
        self.service_manager = service_manager
        self.loop = loop
        self.integ_test = integ_test
        if rpc_stubs is None:
            rpc_stubs = {}
        self.rpc_stubs = rpc_stubs


Controller = namedtuple('Controller', ['name', 'app_future'])


class PipelinedController(Enum):
    InOut = Controller(
        'magma.pipelined.app.inout', 'inout'
    )
    Arp = Controller(
        'magma.pipelined.app.arp', 'arpd'
    )
    Enforcement = Controller(
        'magma.pipelined.app.enforcement', 'enforcement'
    )
    Enforcement_stats = Controller(
        'magma.pipelined.app.enforcement_stats', 'enforcement_stats'
    )
    Testing = Controller(
        'magma.pipelined.app.testing', 'testing'
    )
    Meter = Controller(
        'magma.pipelined.app.meter', 'meter'
    )
    MeterStats = Controller(
        'magma.pipelined.app.meter_stats', 'meter_stats'
    )
    AccessControl = Controller(
        'magma.pipelined.app.access_control', 'access_control'
    )
    Subscriber = Controller(
        'magma.pipelined.app.subscriber', 'subscriber'
    )
    UEMac = Controller(
        'magma.pipelined.app.ue_mac', 'ue_mac'
    )
    TunnelLearnController = Controller(
        'magma.pipelined.app.tunnel_learn', 'tunnel_learn'
    )
    PacketTracer = Controller(
        'magma.pipelined.app.packet_tracer', 'packet_tracer'
    )


def assert_pipelined_not_running():
    """
    As Ryu applications shoudn't be started if the magma@pipelined service is
    running we need to verify if pipelined is active. If service is running
    throws a ServiceRunningError exception.

    This can be done using the command:
        systemctl is-active magma@pipelined
    If service is pipelined, this returns an error code 3 & message "inactive"
    """
    try:
        output = subprocess.check_output(
            ["systemctl", "is-active", "magma@pipelined"]
        )
    except subprocess.CalledProcessError as e:
        if "inactive" not in str(e.output, 'utf-8'):
            raise ServiceRunningError(
                "Pipelined is running, 'systemctl is-active magma@pipelined'" +
                "caused an error code %d, exception - %s"
                % (e.returncode, str(e.output, 'utf-8').strip())
            )
    else:
        raise ServiceRunningError(
            "Pipelined is running, 'systemctl is-active magma@pipelined'" +
            "output - %s" % str(output, 'utf-8').strip()
        )


class StartThread(object):
    """
    Starts ryu applications

    Uses ryu hub and ryu app_manager to launch ryu applications. By using
    futures get references to the instantiated apps. This allows unittests to
    call methods from pipelined apps.
    """
    _Event = namedtuple('_Event', ['func', 'future'])

    def __init__(self, test_setup, launch_successful_future):
        """ If verification fails throw an exception, don't start ryu apps """
        if test_setup.integ_test is False:
            hub.patch(thread=True)
            assert_pipelined_not_running()

        self._test_setup = test_setup
        self.keep_running = True
        self.done = False
        self.event_queue = hub.Queue()
        thread = threading.Thread(
            target=self.start_ryu_apps, args=(launch_successful_future,))
        thread.daemon = True
        thread.start()

    def start_ryu_apps(self, launch_successful_future):
        """
        Starts up ryu applications, all the configuration is parsed from the
        test_setup config provided in the unit test.

        If apps throw an exception on launch, error is passed in the
        launch_successful_future and will prevent infinitely waiting.
        """
        self.reset_static_vars()
        hub.spawn(self._process_queue)

        app_lists = [a.value.name for a in self._test_setup.apps]
        app_futures = {
            controller.value.app_future: future
            for (controller, future) in self._test_setup.references.items()
        }

        manager = AppManager.get_instance()
        manager.load_apps(app_lists)
        contexts = manager.create_contexts()
        contexts['sids_by_ip'] = {}     # shared by both metering apps
        contexts['rule_id_mapper'] = RuleIDToNumMapper()
        contexts['session_rule_version_mapper'] = \
            self._test_setup.service_manager.session_rule_version_mapper
        contexts['app_futures'] = app_futures
        contexts['config'] = self._test_setup.config
        contexts['mconfig'] = self._test_setup.mconfig
        contexts['loop'] = self._test_setup.loop
        contexts['rpc_stubs'] = self._test_setup.rpc_stubs
        contexts['service_manager'] = self._test_setup.service_manager

        logging.basicConfig(
            level=logging.INFO,
            format='[%(asctime)s %(levelname)s %(name)s] %(message)s')

        services = []
        try:
            services.extend(manager.instantiate_apps(**contexts))
        except Exception as e:
            launch_successful_future.set_result(
                "Ryu apps launch exception: {}".format(e))
            raise

        launch_successful_future.set_result("Setup successful")

        self.run(manager)

    def _process_queue(self):
        """
        Run a queue to process external events that need to be run in the Ryu
        greenthread
        """
        while self.keep_running:
            try:
                event = self.event_queue.get(block=False)
                val = event.func()
                event.future.set_result(val)
            except hub.QueueEmpty:
                pass
            finally:
                hub.sleep(0.1)

    def run_in_greenthread(self, func):
        """
        When not monkey patching (i.e. when running a gRPC server), you cannot
        call directly into a Ryu app. To do this, there needs to be a boundary
        between futures and hub.Queues. When this function is called, a lambda
        is passed which is sent into a queue to be run by the Ryu greenthread.
        """
        ev = self._Event(func=func, future=Future())
        self.event_queue.put(ev)
        return ev.future.result()

    def run(self, manager):
        """ Keep running until signalled from test file """
        while self.keep_running:
            hub.sleep(1)

        manager.close()
        self.done = True

    def reset_static_vars(self):
        """ Reset static vars for running nosetests """
        AppManager._instance = AppManager()
        MagmaController.TABLES = {}

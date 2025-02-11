# Copyright (c) Facebook, Inc. and its affiliates.
# All rights reserved.
#
# This source code is licensed under the BSD-style license found in the
# LICENSE file in the root directory of this source tree.
#
###############################################################################
# This file contains common Make targets related to setting up a Python
# environment, running tests, and cleaning up. See lte/gateway/python/Makefile
# for an example of how to use this file.
###############################################################################

# virtualenv bin and build dirs
PYTHON_VERSION=3.5
BIN := $(PYTHON_BUILD)/bin
SRC := $(MAGMA_ROOT)
SITE_PACKAGES_DIR := $(PYTHON_BUILD)/lib/python$(PYTHON_VERSION)/site-packages

# Command to pip install into the virtualenv
VIRT_ENV_PIP_INSTALL := $(BIN)/pip3 install -q -U --cache-dir $(PIP_CACHE_HOME)

install_virtualenv:
	@echo "Initializing virtualenv with python version $(PYTHON_VERSION)"
	virtualenv --system-site-packages --python=/usr/bin/python$(PYTHON_VERSION) $(PYTHON_BUILD)
	. $(PYTHON_BUILD)/bin/activate;
	$(VIRT_ENV_PIP_INSTALL) "pip>=19.1.1"

setupenv: $(PYTHON_BUILD)/sysdeps $(SITE_PACKAGES_DIR)/setuptools

# Sytem packages needed for build
SYS_DEPENDENCIES := python3-dev
$(PYTHON_BUILD)/sysdeps: $(PYTHON_BUILD)
	sudo apt-get -y install $(SYS_DEPENDENCIES)
	touch $(PYTHON_BUILD)/sysdeps

$(PYTHON_BUILD):
	mkdir -p $(PYTHON_BUILD)

$(SITE_PACKAGES_DIR)/setuptools: install_virtualenv
	$(VIRT_ENV_PIP_INSTALL) "setuptools>=41.0.1"

protos:: $(BIN)/grpcio-tools $(PROTO_LIST) prometheus_proto
	@find $(PYTHON_BUILD)/gen -type d | tail -n +2 | sed '/__pycache__/d' | xargs -I % touch "%/__init__.py"
$(PROTO_LIST): %_protos:
	@echo "Generating python code for $* .proto files"
	@mkdir -p $(PYTHON_BUILD)/gen
	@echo "$(PYTHON_BUILD)/gen" > $(SITE_PACKAGES_DIR)/magma_gen.pth
	$(BIN)/python $(SRC)/protos/gen_protos.py $(SRC)/$*/protos/ $(MAGMA_ROOT),$(MAGMA_ROOT)/orc8r/protos/prometheus $(SRC) $(PYTHON_BUILD)/gen/

prometheus_proto:
	$(BIN)/python $(SRC)/protos/gen_prometheus_proto.py $(MAGMA_ROOT) $(PYTHON_BUILD)/gen

# If you update the version here, you probably also want to update it in setup.py
$(BIN)/grpcio-tools: install_virtualenv
	$(VIRT_ENV_PIP_INSTALL) "grpcio-tools==1.25.0"

.test: .tests .sudo_tests

.tests:
ifdef TESTS
	. $(PYTHON_BUILD)/bin/activate; $(BIN)/nosetests --with-coverage --cover-erase --cover-branches --cover-package=magma -s $(TESTS)
endif

.sudo_tests:
ifdef SUDO_TESTS
ifndef SKIP_SUDO_TESTS
	. $(PYTHON_BUILD)/bin/activate; sudo $(BIN)/nosetests --with-coverage --cover-branches --cover-package=magma -s $(SUDO_TESTS)
endif
endif

install_egg: install_virtualenv setup.py
	$(eval NAME ?= $(shell $(BIN)/python setup.py --name))
	@echo "Installing egg link for $(NAME)"
	$(VIRT_ENV_PIP_INSTALL) -e .[dev]

remove_egg:
	rm -rf *.egg-info

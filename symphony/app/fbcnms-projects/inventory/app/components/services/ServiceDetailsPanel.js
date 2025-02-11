/**
 * Copyright 2004-present Facebook. All Rights Reserved.
 *
 * This source code is licensed under the BSD-style license found in the
 * LICENSE file in the root directory of this source tree.
 *
 * @flow
 * @format
 */

import type {Property} from '../../common/Property';
import type {Service} from '../../common/Service';

import ArrowBackIcon from '@material-ui/icons/ArrowBack';
import CustomerTypeahead from '../typeahead/CustomerTypeahead';
import EditServiceMutation from '../../mutations/EditServiceMutation';
import ExpandingPanel from '@fbcnms/ui/components/ExpandingPanel';
import FormField from '@fbcnms/ui/components/design-system/FormField/FormField';
import IconButton from '@material-ui/core/IconButton';
import PropertyValueInput from '../form/PropertyValueInput';
import React, {useRef, useState} from 'react';
import SideBar from '@fbcnms/ui/components/layout/SideBar';
import TextField from '@material-ui/core/TextField';
import symphony from '@fbcnms/ui/theme/symphony';
import update from 'immutability-helper';
import useVerticalScrollingEffect from '../../common/useVerticalScrollingEffect';
import {createFragmentContainer, graphql} from 'react-relay';
import {getInitialPropertyFromType} from '../../common/PropertyType';
import {
  getNonInstancePropertyTypes,
  sortPropertiesByIndex,
  toPropertyInput,
} from '../../common/Property';
import {makeStyles} from '@material-ui/styles';

type Props = {
  shown: boolean,
  service: Service,
  panelWidth?: number,
  onClose: () => void,
};

const useStyles = makeStyles({
  sideBar: {
    border: 'none',
    boxShadow: 'none',
    borderRadius: '0px',
    padding: '0px',
  },
  separator: {
    borderBottom: `1px solid ${symphony.palette.separator}`,
    marginTop: '8px',
  },
  expanded: {
    padding: '0px',
  },
  panel: {
    '&$expanded': {
      margin: '0px',
    },
    boxShadow: 'none',
    padding: '0px',
    background: 'transparent',
  },
  scroller: {
    overflowY: 'auto',
  },
  closeButton: {
    '&&': {
      backgroundColor: symphony.palette.D10,
      color: 'blue',
      margin: '32px 0px 0px 32px',
      padding: '2px',
      display: 'inline-block',
      '&:hover': {
        backgroundColor: symphony.palette.D100,
      },
    },
  },
  expansionPanel: {
    '&&': {
      padding: '24px 20px 16px 32px',
    },
  },
  topBar: {
    display: 'flex',
  },
  detailPane: {
    padding: '0px 32px',
  },
  input: {
    marginBottom: '20px',
  },
});

const ServiceDetailsPanel = (props: Props) => {
  const classes = useStyles();
  const {shown, service, panelWidth, onClose} = props;
  const thisElement = useRef(null);
  const [isDirty, setIsDirty] = useState(false);
  useVerticalScrollingEffect(thisElement);
  let properties = service?.properties ?? [];
  if (service.serviceType.propertyTypes) {
    properties = [
      ...properties,
      ...getNonInstancePropertyTypes(
        properties,
        service.serviceType.propertyTypes,
      ).map(propType => getInitialPropertyFromType(propType)),
    ];
    properties = properties.sort(sortPropertiesByIndex);
  }

  const [editableService, setEditableService] = useState({
    id: service.id,
    name: service.name,
    externalId: service.externalId,
    customer: service.customer,
    properties: properties,
  });

  const getServiceInput = () => {
    return {
      data: {
        id: editableService.id,
        name: editableService.name,
        externalId: editableService.externalId,
        customerId: editableService.customer?.id,
        upstreamServiceIds: [],
        properties: toPropertyInput(editableService.properties),
        terminationPointIds: [],
      },
    };
  };

  const onChangeProperty = index => (property: Property) => {
    setEditableService(
      update(editableService, {
        properties: {
          [index]: {$set: property},
        },
      }),
    );
    setIsDirty(true);
  };
  const onChangeDetail = (key: 'name' | 'externalId' | 'customer', value) => {
    // $FlowFixMe Update specific value
    setEditableService(update(editableService, {[key]: {$set: value}}));
    setIsDirty(true);
  };

  const onBlur = () => {
    if (isDirty) {
      EditServiceMutation(getServiceInput());
    }
  };

  const backButton = (props: {onClose: () => void}) => (
    <div className={classes.topBar}>
      <IconButton className={classes.closeButton} onClick={props.onClose}>
        <ArrowBackIcon fontSize="small" color="primary" />
      </IconButton>
    </div>
  );
  return (
    <SideBar
      isShown={shown}
      top={0}
      width={panelWidth}
      onClose={onClose}
      className={classes.sideBar}
      backButton={backButton}>
      <div ref={thisElement} className={classes.scroller}>
        <ExpandingPanel
          title="Details"
          defaultExpanded={true}
          expandedClassName={classes.expanded}
          expansionPanelSummaryClassName={classes.expansionPanel}
          detailsPaneClass={classes.detailPane}
          className={classes.panel}>
          <div className={classes.input}>
            <FormField label="Name">
              <TextField
                name="name"
                variant="outlined"
                margin="dense"
                onChange={event => onChangeDetail('name', event.target.value)}
                value={editableService.name}
                onBlur={onBlur}
              />
            </FormField>
          </div>
          <div className={classes.input}>
            <FormField label="Service ID">
              <TextField
                name="serviceId"
                variant="outlined"
                margin="dense"
                onChange={event =>
                  onChangeDetail('externalId', event.target.value)
                }
                value={editableService.externalId ?? ''}
                onBlur={onBlur}
              />
            </FormField>
          </div>
          <div className={classes.input}>
            <FormField label="Service Type">
              <TextField
                disabled
                name="type"
                variant="outlined"
                margin="dense"
                value={service.serviceType.name}
              />
            </FormField>
          </div>
          <div className={classes.input}>
            <FormField label="Customer">
              <CustomerTypeahead
                onCustomerSelection={customer => {
                  onChangeDetail('customer', customer);
                  onBlur();
                }}
                required={false}
                selectedCustomer={editableService.customer?.name}
                margin="dense"
              />
            </FormField>
          </div>
        </ExpandingPanel>
        <div className={classes.separator} />
        <ExpandingPanel
          title="Properties"
          defaultExpanded={true}
          expandedClassName={classes.expanded}
          expansionPanelSummaryClassName={classes.expansionPanel}
          detailsPaneClass={classes.detailPane}
          className={classes.panel}>
          {editableService.properties.map((property, index) => (
            <PropertyValueInput
              fullWidth
              required={!!property.propertyType.isInstanceProperty}
              disabled={!property.propertyType.isInstanceProperty}
              label={property.propertyType.name}
              className={classes.input}
              margin="dense"
              inputType="Property"
              property={property}
              // $FlowFixMe pass property and not property type
              onChange={onChangeProperty(index)}
              onBlur={onBlur}
              headlineVariant="form"
            />
          ))}
        </ExpandingPanel>
      </div>
    </SideBar>
  );
};

export default createFragmentContainer(ServiceDetailsPanel, {
  service: graphql`
    fragment ServiceDetailsPanel_service on Service {
      id
      name
      externalId
      customer {
        name
      }
      serviceType {
        id
        name
        propertyTypes {
          id
          name
          index
          isInstanceProperty
          type
          stringValue
          intValue
          floatValue
          booleanValue
          latitudeValue
          longitudeValue
          rangeFromValue
          rangeToValue
        }
      }
      properties {
        id
        propertyType {
          id
          name
          type
          isEditable
          isInstanceProperty
          stringValue
        }
        stringValue
        intValue
        floatValue
        booleanValue
        latitudeValue
        longitudeValue
        rangeFromValue
        rangeToValue
        equipmentValue {
          id
          name
        }
        locationValue {
          id
          name
        }
      }
    }
  `,
});

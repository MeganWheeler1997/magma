// Copyright (c) 2004-present Facebook All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package exporter

import (
	"context"
	"flag"
	"os"
	"testing"

	"github.com/AlekSi/pointer"
	"github.com/facebookincubator/symphony/cloud/log/logtest"
	"github.com/facebookincubator/symphony/cloud/testdb"
	"github.com/facebookincubator/symphony/graph/ent"
	"github.com/facebookincubator/symphony/graph/ent/equipmentportdefinition"
	"github.com/facebookincubator/symphony/graph/ent/equipmentpositiondefinition"
	"github.com/facebookincubator/symphony/graph/ent/propertytype"
	"github.com/facebookincubator/symphony/graph/graphql/generated"
	"github.com/facebookincubator/symphony/graph/graphql/models"
	"github.com/facebookincubator/symphony/graph/graphql/resolver"

	"github.com/facebookincubator/ent/dialect"
	"github.com/facebookincubator/ent/dialect/sql"
	"github.com/facebookincubator/ent/dialect/sql/schema"
	"github.com/stretchr/testify/require"
)

var debug = flag.Bool("debug", false, "run database driver on debug mode")

const (
	tenantHeader        = "x-auth-organization"
	equipmentTypeName   = "equipmentType"
	equipmentType2Name  = "equipmentType2"
	parentEquip         = "parentEquipmentName"
	currEquip           = "currEquipmentName"
	currEquip2          = "currEquipmentName2"
	positionName        = "Position"
	portName1           = "port1"
	portName2           = "port2"
	portName3           = "port3"
	propNameStr         = "propNameStr"
	propNameInt         = "propNameInts"
	newPropNameStr      = "newPropNameStr"
	propDefValue        = "defaultVal"
	propDefValue2       = "defaultVal2"
	propDevValInt       = 15
	propInstanceValue   = "newVal"
	locTypeNameL        = "locTypeLarge"
	locTypeNameM        = "locTypeMedium"
	locTypeNameS        = "locTypeSmall"
	grandParentLocation = "grandParentLocation"
	parentLocation      = "parentLocation"
	childLocation       = "childLocation"
)

func TestMain(m *testing.M) {
	flag.Parse()
	os.Exit(m.Run())
}

type TestExporterResolver struct {
	generated.ResolverRoot
	drv      dialect.Driver
	client   *ent.Client
	exporter exporter
}

func newExporterTestResolver(t *testing.T) (*TestExporterResolver, error) {
	db, name, err := testdb.Open()
	require.NoError(t, err)
	db.SetMaxOpenConns(1)
	return newResolver(t, sql.OpenDB(name, db))
}

func newResolver(t *testing.T, drv dialect.Driver) (*TestExporterResolver, error) {
	if *debug {
		drv = dialect.Debug(drv)
	}
	client := ent.NewClient(ent.Driver(drv))
	require.NoError(t, client.Schema.Create(context.Background(), schema.WithGlobalUniqueID(true)))

	r, err := resolver.New(logtest.NewTestLogger(t))
	require.NoError(t, err)
	log := logtest.NewTestLogger(t)

	e := exporter{log, equipmentRower{log}}
	if err != nil {
		return nil, err
	}
	return &TestExporterResolver{r, drv, client, e}, nil
}

func prepareData(ctx context.Context, t *testing.T, r TestExporterResolver) {
	mr := r.Mutation()

	locTypeL, err := mr.AddLocationType(ctx, models.AddLocationTypeInput{Name: locTypeNameL})
	require.NoError(t, err)
	locTypeM, err := mr.AddLocationType(ctx, models.AddLocationTypeInput{Name: locTypeNameM})
	require.NoError(t, err)
	locTypeS, err := mr.AddLocationType(ctx, models.AddLocationTypeInput{Name: locTypeNameS})
	require.NoError(t, err)

	_, err = mr.EditLocationTypesIndex(ctx, []*models.LocationTypeIndex{
		{
			LocationTypeID: locTypeL.ID,
			Index:          0,
		},
		{
			LocationTypeID: locTypeM.ID,
			Index:          1,
		},
		{
			LocationTypeID: locTypeS.ID,
			Index:          2,
		},
	})
	require.NoError(t, err)

	gpLocation, err := mr.AddLocation(ctx, models.AddLocationInput{
		Name: grandParentLocation,
		Type: locTypeL.ID,
	})
	require.NoError(t, err)
	pLocation, err := mr.AddLocation(ctx, models.AddLocationInput{
		Name:   parentLocation,
		Type:   locTypeM.ID,
		Parent: &gpLocation.ID,
	})
	require.NoError(t, err)
	clocation, err := mr.AddLocation(ctx, models.AddLocationInput{
		Name:   childLocation,
		Type:   locTypeS.ID,
		Parent: &pLocation.ID,
	})
	require.NoError(t, err)
	position1 := models.EquipmentPositionInput{
		Name: positionName,
	}

	ptyp, _ := mr.AddEquipmentPortType(ctx, models.AddEquipmentPortTypeInput{
		Name: "portType1",
		Properties: []*models.PropertyTypeInput{
			{
				Name:        propStr,
				Type:        "string",
				StringValue: pointer.ToString("t1"),
			},
			{
				Name: propStr2,
				Type: "string",
			},
		},
	})
	port1 := models.EquipmentPortInput{
		Name:       portName1,
		PortTypeID: &ptyp.ID,
	}
	strDefVal := propDefValue
	intDefVal := propDevValInt
	propDefInput1 := models.PropertyTypeInput{
		Name:        propNameStr,
		Type:        "string",
		StringValue: &strDefVal,
	}
	propDefInput2 := models.PropertyTypeInput{
		Name:     propNameInt,
		Type:     "int",
		IntValue: &intDefVal,
	}
	equipmentType, err := mr.AddEquipmentType(ctx, models.AddEquipmentTypeInput{
		Name:      equipmentTypeName,
		Positions: []*models.EquipmentPositionInput{&position1},
		Ports:     []*models.EquipmentPortInput{&port1},
	})
	require.NoError(t, err)

	port2 := models.EquipmentPortInput{
		Name: portName2,
	}
	equipmentType2, err := mr.AddEquipmentType(ctx, models.AddEquipmentTypeInput{
		Name:       equipmentType2Name,
		Properties: []*models.PropertyTypeInput{&propDefInput1, &propDefInput2},
		Ports:      []*models.EquipmentPortInput{&port2},
	})
	require.NoError(t, err)

	posDef1 := equipmentType.QueryPositionDefinitions().Where(equipmentpositiondefinition.Name(positionName)).OnlyX(ctx)
	propDef1 := equipmentType2.QueryPropertyTypes().Where(propertytype.Name(propNameStr)).OnlyX(ctx)

	parentEquipment, err := mr.AddEquipment(ctx, models.AddEquipmentInput{
		Name:     parentEquip,
		Type:     equipmentType.ID,
		Location: &clocation.ID,
	})
	require.NoError(t, err)

	strVal := propInstanceValue
	propInstance1 := models.PropertyInput{
		PropertyTypeID: propDef1.ID,
		StringValue:    &strVal,
	}
	childEquip, err := mr.AddEquipment(ctx, models.AddEquipmentInput{
		Name:               currEquip,
		Type:               equipmentType2.ID,
		Parent:             &parentEquipment.ID,
		PositionDefinition: &posDef1.ID,
		Properties:         []*models.PropertyInput{&propInstance1},
	})
	require.NoError(t, err)

	portDef1 := equipmentType.QueryPortDefinitions().Where(equipmentportdefinition.Name(portName1)).OnlyX(ctx)
	portDef2 := equipmentType2.QueryPortDefinitions().Where(equipmentportdefinition.Name(portName2)).OnlyX(ctx)
	_, _ = mr.AddLink(ctx, models.AddLinkInput{
		Sides: []*models.LinkSide{
			{Equipment: parentEquipment.ID, Port: portDef1.ID},
			{Equipment: childEquip.ID, Port: portDef2.ID},
		},
	})

	val := propDefValue2
	propertyInput := models.PropertyTypeInput{Name: newPropNameStr, StringValue: &val, Type: models.PropertyKindString}
	_, err = r.Mutation().EditEquipmentType(ctx, models.EditEquipmentTypeInput{
		ID:         equipmentType2.ID,
		Name:       equipmentType2.Name,
		Properties: []*models.PropertyTypeInput{&propertyInput},
	})
	require.NoError(t, err)
}

// Copyright (c) 2004-present Facebook All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package resolver

import (
	"context"
	"testing"

	"github.com/facebookincubator/symphony/graph/ent/propertytype"

	"github.com/AlekSi/pointer"

	"github.com/facebookincubator/symphony/graph/graphql/models"
	"github.com/facebookincubator/symphony/graph/viewer/viewertest"

	"github.com/stretchr/testify/require"
)

type linkSearchDataModels struct {
	e1   string
	e2   string
	e3   string
	e4   string
	loc1 string
	l1   string
	l2   string
}

type linkSearchHirerchyDataModels struct {
	e1 string
	e2 string
	e3 string
	e4 string
	e5 string
	e6 string
}

func prepareLinkData(ctx context.Context, r *TestResolver, props []*models.PropertyInput) linkSearchDataModels {
	mr := r.Mutation()
	wot, _ := mr.AddWorkOrderType(ctx, models.AddWorkOrderTypeInput{Name: "WO-type1"})
	wo1, _ := mr.AddWorkOrder(ctx, models.AddWorkOrderInput{Name: "wo1", WorkOrderTypeID: wot.ID})
	wo2, _ := mr.AddWorkOrder(ctx, models.AddWorkOrderInput{Name: "wo2", WorkOrderTypeID: wot.ID})
	wo2, _ = mr.EditWorkOrder(ctx, models.EditWorkOrderInput{ID: wo2.ID, Name: "wo2", Status: models.WorkOrderStatusDone})
	locType1, _ := mr.AddLocationType(ctx, models.AddLocationTypeInput{
		Name: "loc_type1",
	})

	loc1, _ := mr.AddLocation(ctx, models.AddLocationInput{
		Name: "loc_inst1",
		Type: locType1.ID,
	})

	ptyp, _ := mr.AddEquipmentPortType(ctx, models.AddEquipmentPortTypeInput{
		Name: "portType1",
		LinkProperties: []*models.PropertyTypeInput{
			{
				Name:        "propStr",
				Type:        "string",
				StringValue: pointer.ToString("t1"),
			},
			{
				Name: "connected_date",
				Type: models.PropertyKindDate,
			},
		},
	})

	equType, _ := mr.AddEquipmentType(ctx, models.AddEquipmentTypeInput{
		Name: "eq_type",
		Ports: []*models.EquipmentPortInput{
			{Name: "typ1_p1", PortTypeID: &ptyp.ID},
			{Name: "typ1_p2"},
		},
	})
	defs := equType.QueryPortDefinitions().AllX(ctx)
	equType2, _ := mr.AddEquipmentType(ctx, models.AddEquipmentTypeInput{
		Name: "eq_type2",
		Ports: []*models.EquipmentPortInput{
			{Name: "typ2_p1"},
			{Name: "typ2_p2"},
		},
	})
	defs2 := equType2.QueryPortDefinitions().AllX(ctx)

	e1, _ := mr.AddEquipment(ctx, models.AddEquipmentInput{
		Name:       "eq_inst1",
		Type:       equType.ID,
		Location:   &loc1.ID,
		Properties: props,
	})
	e2, _ := mr.AddEquipment(ctx, models.AddEquipmentInput{
		Name:       "eq_inst2",
		Type:       equType.ID,
		Location:   &loc1.ID,
		Properties: props,
	})
	e3, _ := mr.AddEquipment(ctx, models.AddEquipmentInput{
		Name:       "eq_inst3",
		Type:       equType2.ID,
		Location:   &loc1.ID,
		Properties: props,
	})
	e4, _ := mr.AddEquipment(ctx, models.AddEquipmentInput{
		Name:       "eq_inst4",
		Type:       equType2.ID,
		Location:   &loc1.ID,
		Properties: props,
	})

	strProp := ptyp.QueryLinkPropertyTypes().Where(propertytype.Name("propStr")).OnlyX(ctx)
	dateProp := ptyp.QueryLinkPropertyTypes().Where(propertytype.Name("connected_date")).OnlyX(ctx)

	l1, _ := mr.AddLink(ctx, models.AddLinkInput{
		Sides: []*models.LinkSide{
			{Equipment: e1.ID, Port: defs[0].ID},
			{Equipment: e3.ID, Port: defs2[0].ID},
		},
		Properties: []*models.PropertyInput{
			{
				PropertyTypeID: strProp.ID,
				StringValue:    pointer.ToString("newVal"),
			},
			{
				PropertyTypeID: dateProp.ID,
				StringValue:    pointer.ToString("1988-03-29"),
			},
		},
		WorkOrder: &wo1.ID,
	})
	l2, _ := mr.AddLink(ctx, models.AddLinkInput{
		Sides: []*models.LinkSide{
			{Equipment: e2.ID, Port: defs[1].ID},
			{Equipment: e4.ID, Port: defs2[1].ID},
		},
		WorkOrder: &wo2.ID,
	})
	_, _ = mr.RemoveLink(ctx, l2.ID, &wo2.ID)
	return linkSearchDataModels{
		e1.ID,
		e2.ID,
		e3.ID,
		e4.ID,
		loc1.ID,
		l1.ID,
		l2.ID,
	}
}

func prepareLinkDataByHirerchy(ctx context.Context, r *TestResolver) linkSearchHirerchyDataModels {
	mr := r.Mutation()
	locType, _ := mr.AddLocationType(ctx, models.AddLocationTypeInput{
		Name: "loc_type",
	})

	loc, _ := mr.AddLocation(ctx, models.AddLocationInput{
		Name: "loc_inst",
		Type: locType.ID,
	})

	equType, _ := mr.AddEquipmentType(ctx, models.AddEquipmentTypeInput{
		Name: "eq_type",
		Ports: []*models.EquipmentPortInput{
			{Name: "typ1_port1"},
			{Name: "typ1_port2"},
		},
		Positions: []*models.EquipmentPositionInput{
			{Name: "typ1_pos1"},
			{Name: "typ1_pos2"},
		},
	})
	posDefs := equType.QueryPositionDefinitions().AllX(ctx)
	portDefs := equType.QueryPortDefinitions().AllX(ctx)

	equType2, _ := mr.AddEquipmentType(ctx, models.AddEquipmentTypeInput{
		Name: "eq_type2",
		Ports: []*models.EquipmentPortInput{
			{Name: "typ2_p1"},
			{Name: "typ2_p2"},
		},
	})
	portDefs2 := equType2.QueryPortDefinitions().AllX(ctx)

	e1, _ := mr.AddEquipment(ctx, models.AddEquipmentInput{
		Name:     "eq_inst1",
		Type:     equType.ID,
		Location: &loc.ID,
	})
	e2, _ := mr.AddEquipment(ctx, models.AddEquipmentInput{
		Name:     "eq_inst2",
		Type:     equType.ID,
		Location: &loc.ID,
	})
	e3, _ := mr.AddEquipment(ctx, models.AddEquipmentInput{
		Name:               "eq_inst3",
		Type:               equType2.ID,
		Parent:             &e1.ID,
		PositionDefinition: &posDefs[0].ID,
	})
	e4, _ := mr.AddEquipment(ctx, models.AddEquipmentInput{
		Name:               "eq_inst4",
		Type:               equType2.ID,
		Parent:             &e1.ID,
		PositionDefinition: &posDefs[1].ID,
	})
	e5, _ := mr.AddEquipment(ctx, models.AddEquipmentInput{
		Name:               "eq_inst5",
		Type:               equType2.ID,
		Parent:             &e2.ID,
		PositionDefinition: &posDefs[0].ID,
	})
	e6, _ := mr.AddEquipment(ctx, models.AddEquipmentInput{
		Name:               "eq_inst6",
		Type:               equType2.ID,
		Parent:             &e2.ID,
		PositionDefinition: &posDefs[1].ID,
	})

	_, _ = mr.AddLink(ctx, models.AddLinkInput{
		Sides: []*models.LinkSide{
			{Equipment: e1.ID, Port: portDefs[0].ID},
			{Equipment: e2.ID, Port: portDefs[0].ID},
		},
	})
	_, _ = mr.AddLink(ctx, models.AddLinkInput{
		Sides: []*models.LinkSide{
			{Equipment: e1.ID, Port: portDefs[1].ID},
			{Equipment: e5.ID, Port: portDefs2[0].ID},
		},
	})

	_, _ = mr.AddLink(ctx, models.AddLinkInput{
		Sides: []*models.LinkSide{
			{Equipment: e2.ID, Port: portDefs[1].ID},
			{Equipment: e3.ID, Port: portDefs2[0].ID},
		},
	})

	_, _ = mr.AddLink(ctx, models.AddLinkInput{
		Sides: []*models.LinkSide{
			{Equipment: e4.ID, Port: portDefs[0].ID},
			{Equipment: e6.ID, Port: portDefs2[0].ID},
		},
	})

	return linkSearchHirerchyDataModels{
		e1.ID,
		e2.ID,
		e3.ID,
		e4.ID,
		e5.ID,
		e6.ID,
	}
}

func TestSearchLinksFutureState(t *testing.T) {
	r, err := newTestResolver(t)
	require.NoError(t, err)
	defer r.drv.Close()
	ctx := viewertest.NewContext(r.client)

	data := prepareLinkData(ctx, r, nil)
	/*
		helper: data now is of type:
		(loc1) link1 :
			e1(pos1, type1) <--> e3 (pos1, type2)
			state: PENDING
		(loc1) link2 :
			e2(pos2, type1) <--> e4 (pos2, type2)
			state: DONE
	*/
	qr := r.Query()
	limit := 100
	all, err := qr.LinkSearch(ctx, []*models.LinkFilterInput{}, &limit)
	require.NoError(t, err)
	require.Len(t, all.Links, 2)
	require.Equal(t, all.Count, 2)
	maxDepth := 2
	f1 := models.LinkFilterInput{
		FilterType: models.LinkFilterTypeLinkFutureStatus,
		Operator:   models.FilterOperatorIsOneOf,
		IDSet:      []string{models.FutureStateRemove.String()},
		MaxDepth:   &maxDepth,
	}
	res1, err := qr.LinkSearch(ctx, []*models.LinkFilterInput{&f1}, &limit)
	require.NoError(t, err)
	require.Len(t, res1.Links, 1)
	ports := res1.Links[0].QueryPorts().AllX(ctx)
	require.NotEqual(t, ports[0].QueryParent().OnlyX(ctx).ID, ports[1].QueryParent().OnlyX(ctx).ID)
	for _, port := range ports {
		prnt := port.QueryParent().OnlyX(ctx).ID
		require.Contains(t, []string{data.e2, data.e4}, prnt)
	}
}

func TestSearchLinksByLocation(t *testing.T) {
	r, err := newTestResolver(t)
	require.NoError(t, err)
	defer r.drv.Close()
	ctx := viewertest.NewContext(r.client)

	data := prepareLinkData(ctx, r, nil)
	/*
		helper: data now is of type:
		(loc1) link1 :
			e1(pos1, type1) <--> e3 (pos1, type2)
			state: PENDING
		(loc1) link2 :
			e2(pos2, type1) <--> e4 (pos2, type2)
			state: DONE
	*/
	qr, mr := r.Query(), r.Mutation()
	limit := 100
	all, err := qr.LinkSearch(ctx, []*models.LinkFilterInput{}, &limit)
	require.NoError(t, err)
	require.Len(t, all.Links, 2)
	maxDepth := 2
	f1 := models.LinkFilterInput{
		FilterType: models.LinkFilterTypeLocationInst,
		Operator:   models.FilterOperatorIsOneOf,
		IDSet:      []string{data.loc1},
		MaxDepth:   &maxDepth,
	}
	typ, _ := mr.AddLocationType(ctx, models.AddLocationTypeInput{
		Name: "loc_t",
	})

	loc, _ := mr.AddLocation(ctx, models.AddLocationInput{
		Name: "loc",
		Type: typ.ID,
	})
	res1, err := qr.LinkSearch(ctx, []*models.LinkFilterInput{&f1}, &limit)
	require.NoError(t, err)
	require.Len(t, res1.Links, 2)
	f2 := models.LinkFilterInput{
		FilterType: models.LinkFilterTypeLocationInst,
		Operator:   models.FilterOperatorIsOneOf,
		IDSet:      []string{loc.ID},
		MaxDepth:   &maxDepth,
	}
	res2, err := qr.LinkSearch(ctx, []*models.LinkFilterInput{&f2}, &limit)
	require.NoError(t, err)
	require.Len(t, res2.Links, 0)
}

func TestSearchLinksByEquipmentTyp(t *testing.T) {
	r, err := newTestResolver(t)
	require.NoError(t, err)
	defer r.drv.Close()
	ctx := viewertest.NewContext(r.client)

	data := prepareLinkData(ctx, r, nil)
	/*
		helper: data now is of type:
		(loc1) link1 :
			e1(pos1, type1) <--> e3 (pos1, type2)
			state: PENDING
		(loc1) link2 :
			e2(pos2, type1) <--> e4 (pos2, type2)
			state: DONE
	*/
	qr, mr := r.Query(), r.Mutation()
	limit := 100
	all, err := qr.LinkSearch(ctx, []*models.LinkFilterInput{}, &limit)
	require.NoError(t, err)
	require.Len(t, all.Links, 2)
	maxDepth := 2
	e1, _ := qr.Equipment(ctx, data.e1)
	typ1 := e1.QueryType().OnlyX(ctx)
	e3, _ := qr.Equipment(ctx, data.e3)
	typ2 := e3.QueryType().OnlyX(ctx)

	emptyTyp, _ := mr.AddEquipmentType(ctx, models.AddEquipmentTypeInput{Name: "empty_typ"})
	f1 := models.LinkFilterInput{
		FilterType: models.LinkFilterTypeEquipmentType,
		Operator:   models.FilterOperatorIsOneOf,
		IDSet:      []string{typ1.ID},
		MaxDepth:   &maxDepth,
	}
	res1, err := qr.LinkSearch(ctx, []*models.LinkFilterInput{&f1}, &limit)
	require.NoError(t, err)
	require.Len(t, res1.Links, 2)

	f2 := models.LinkFilterInput{
		FilterType: models.LinkFilterTypeEquipmentType,
		Operator:   models.FilterOperatorIsOneOf,
		IDSet:      []string{typ2.ID},
		MaxDepth:   &maxDepth,
	}
	res2, err := qr.LinkSearch(ctx, []*models.LinkFilterInput{&f2}, &limit)
	require.NoError(t, err)
	require.Len(t, res2.Links, 2)

	f3 := models.LinkFilterInput{
		FilterType: models.LinkFilterTypeLocationInst,
		Operator:   models.FilterOperatorIsOneOf,
		IDSet:      []string{emptyTyp.ID},
		MaxDepth:   &maxDepth,
	}
	res3, err := qr.LinkSearch(ctx, []*models.LinkFilterInput{&f3}, &limit)
	require.NoError(t, err)
	require.Len(t, res3.Links, 0)
}

func TestSearchLinksByEquipment(t *testing.T) {
	r, err := newTestResolver(t)
	require.NoError(t, err)
	defer r.drv.Close()
	ctx := viewertest.NewContext(r.client)

	data := prepareLinkData(ctx, r, nil)
	/*
		helper: data now is of type:
		(loc1) link1 :
			e1(pos1, type1) <--> e3 (pos1, type2)
			state: PENDING
		(loc1) link2 :
			e2(pos2, type1) <--> e4 (pos2, type2)
			state: DONE
	*/
	qr := r.Query()
	limit := 100
	all, err := qr.LinkSearch(ctx, []*models.LinkFilterInput{}, &limit)
	require.NoError(t, err)
	require.Len(t, all.Links, 2)
	maxDepth := 2

	f1 := models.LinkFilterInput{
		FilterType: models.LinkFilterTypeEquipmentInst,
		Operator:   models.FilterOperatorIsOneOf,
		IDSet:      []string{data.e1, data.e2},
		MaxDepth:   &maxDepth,
	}
	res1, err := qr.LinkSearch(ctx, []*models.LinkFilterInput{&f1}, &limit)
	require.NoError(t, err)
	require.Len(t, res1.Links, 2)

	f2 := models.LinkFilterInput{
		FilterType: models.LinkFilterTypeEquipmentInst,
		Operator:   models.FilterOperatorIsOneOf,
		IDSet:      []string{data.e2, data.e4},
		MaxDepth:   &maxDepth,
	}
	res2, err := qr.LinkSearch(ctx, []*models.LinkFilterInput{&f2}, &limit)
	require.NoError(t, err)
	require.Len(t, res2.Links, 1)
}

func TestSearchLinksByEquipmentHirerchy(t *testing.T) {
	r, err := newTestResolver(t)
	require.NoError(t, err)
	defer r.drv.Close()
	ctx := viewertest.NewContext(r.client)

	data := prepareLinkDataByHirerchy(ctx, r)
	/*
		helper: data now is of type:
			equipments
				e1(pos1) -> e3
				  (pos2) -> e4
				e2(pos1) -> e5
				  (pos2) -> e6
			links
				e1(port1) <--> e2(port1)
				e1(port2) <--> e5(port1)
				e2(port2) <--> e3(port1)
				e4(port1) <--> e6(port1)
	*/

	qr := r.Query()
	limit := 100
	all, err := qr.LinkSearch(ctx, []*models.LinkFilterInput{}, &limit)
	require.NoError(t, err)
	require.Len(t, all.Links, 4)
	maxDepth := 2

	f1 := models.LinkFilterInput{
		FilterType: models.LinkFilterTypeEquipmentInst,
		Operator:   models.FilterOperatorIsOneOf,
		IDSet:      []string{data.e1},
		MaxDepth:   &maxDepth,
	}
	res1, err := qr.LinkSearch(ctx, []*models.LinkFilterInput{&f1}, &limit)
	require.NoError(t, err)
	require.Len(t, res1.Links, 4)

	f2 := models.LinkFilterInput{
		FilterType: models.LinkFilterTypeEquipmentInst,
		Operator:   models.FilterOperatorIsOneOf,
		IDSet:      []string{data.e6},
		MaxDepth:   &maxDepth,
	}
	res2, err := qr.LinkSearch(ctx, []*models.LinkFilterInput{&f2}, &limit)
	require.NoError(t, err)
	require.Len(t, res2.Links, 1)
}

func TestSearchLinksByService(t *testing.T) {
	r, err := newTestResolver(t)
	require.NoError(t, err)
	defer r.drv.Close()
	ctx := viewertest.NewContext(r.client)

	data := prepareLinkData(ctx, r, nil)
	/*
		helper: data now is of type:
		(loc1) link1 :
			e1(pos1, type1) <--> e3 (pos1, type2)
			state: PENDING
		(loc1) link2 :
			e2(pos2, type1) <--> e4 (pos2, type2)
			state: DONE
	*/

	qr, mr := r.Query(), r.Mutation()

	st, _ := mr.AddServiceType(ctx, models.ServiceTypeCreateData{
		Name: "Internet Access", HasCustomer: false})

	s1, err := mr.AddService(ctx, models.ServiceCreateData{
		Name:                "Internet Access Room 2a",
		ServiceTypeID:       st.ID,
		TerminationPointIds: []string{},
	})
	require.NoError(t, err)
	_, err = mr.AddServiceLink(ctx, s1.ID, data.l1)
	require.NoError(t, err)

	s2, err := mr.AddService(ctx, models.ServiceCreateData{
		Name:                "Internet Access Room 2b",
		ServiceTypeID:       st.ID,
		TerminationPointIds: []string{},
	})
	require.NoError(t, err)
	_, err = mr.AddServiceLink(ctx, s2.ID, data.l1)
	require.NoError(t, err)
	_, err = mr.AddServiceLink(ctx, s2.ID, data.l2)
	require.NoError(t, err)

	limit := 100
	all, err := qr.LinkSearch(ctx, []*models.LinkFilterInput{}, &limit)
	require.NoError(t, err)
	require.Len(t, all.Links, 2)
	maxDepth := 2

	f1 := models.LinkFilterInput{
		FilterType: models.LinkFilterTypeServiceInst,
		Operator:   models.FilterOperatorIsOneOf,
		IDSet:      []string{s1.ID},
		MaxDepth:   &maxDepth,
	}
	res1, err := qr.LinkSearch(ctx, []*models.LinkFilterInput{&f1}, &limit)
	require.NoError(t, err)
	require.Len(t, res1.Links, 1)

	f2 := models.LinkFilterInput{
		FilterType: models.LinkFilterTypeServiceInst,
		Operator:   models.FilterOperatorIsOneOf,
		IDSet:      []string{s2.ID},
		MaxDepth:   &maxDepth,
	}
	res2, err := qr.LinkSearch(ctx, []*models.LinkFilterInput{&f2}, &limit)
	require.NoError(t, err)
	require.Len(t, res2.Links, 2)

	f3 := models.LinkFilterInput{
		FilterType: models.LinkFilterTypeServiceInst,
		Operator:   models.FilterOperatorIsNotOneOf,
		IDSet:      []string{s1.ID},
		MaxDepth:   &maxDepth,
	}
	res3, err := qr.LinkSearch(ctx, []*models.LinkFilterInput{&f3}, &limit)
	require.NoError(t, err)
	require.Len(t, res3.Links, 1)

	f4 := models.LinkFilterInput{
		FilterType: models.LinkFilterTypeServiceInst,
		Operator:   models.FilterOperatorIsNotOneOf,
		IDSet:      []string{s2.ID},
		MaxDepth:   &maxDepth,
	}
	res4, err := qr.LinkSearch(ctx, []*models.LinkFilterInput{&f4}, &limit)
	require.NoError(t, err)
	require.Len(t, res4.Links, 0)

}

func TestSearchLinksByProperty(t *testing.T) {
	r, err := newTestResolver(t)
	require.NoError(t, err)
	defer r.drv.Close()
	ctx := viewertest.NewContext(r.client)

	prepareLinkData(ctx, r, nil)
	/*
		helper: data now is of type:
		(loc1) link1 :
			e1(pos1, type1) <--> e3 (pos1, type2)
			state: PENDING
		(loc1) link2 :
			e2(pos2, type1) <--> e4 (pos2, type2)
			state: DONE
	*/
	qr := r.Query()
	limit := 100

	f1 := models.LinkFilterInput{
		FilterType: models.LinkFilterTypeProperty,
		Operator:   models.FilterOperatorIs,
		PropertyValue: &models.PropertyTypeInput{
			Name:        "propStr",
			Type:        models.PropertyKindString,
			StringValue: pointer.ToString("newVal"),
		},
	}

	res1, err := qr.LinkSearch(ctx, []*models.LinkFilterInput{&f1}, &limit)
	require.NoError(t, err)
	links := res1.Links
	require.Len(t, links, 1)

	f2 := models.LinkFilterInput{
		FilterType: models.LinkFilterTypeProperty,
		Operator:   models.FilterOperatorDateLessThan,
		PropertyValue: &models.PropertyTypeInput{
			Name:        "connected_date",
			Type:        models.PropertyKindDate,
			StringValue: pointer.ToString("2019-01-01"),
		},
	}

	res2, err := qr.LinkSearch(ctx, []*models.LinkFilterInput{&f2}, &limit)
	require.NoError(t, err)
	links = res2.Links
	require.Len(t, links, 1)
}

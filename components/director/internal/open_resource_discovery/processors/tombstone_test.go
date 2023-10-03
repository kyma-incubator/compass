package processors

//successfulTombstoneCreate := func() *automockproc.TombstoneService {
//	tombstoneSvc := &automockproc.TombstoneService{}
//	tombstoneSvc.On("ListByApplicationID", txtest.CtxWithDBMatcher(), appID).Return(nil, nil).Once()
//	tombstoneSvc.On("Create", txtest.CtxWithDBMatcher(), resource.Application, appID, *sanitizedDoc.Tombstones[0]).Return("", nil).Once()
//	tombstoneSvc.On("ListByApplicationID", txtest.CtxWithDBMatcher(), appID).Return(fixTombstones(), nil).Once()
//	return tombstoneSvc
//}
//
//successfulTombstoneCreateForStaticDoc := func() *automockproc.TombstoneService {
//	tombstoneSvc := &automockproc.TombstoneService{}
//	tombstoneSvc.On("ListByApplicationTemplateVersionID", txtest.CtxWithDBMatcher(), appTemplateVersionID).Return(nil, nil).Once()
//	tombstoneSvc.On("Create", txtest.CtxWithDBMatcher(), resource.ApplicationTemplateVersion, appTemplateVersionID, *sanitizedStaticDoc.Tombstones[0]).Return("", nil).Once()
//	tombstoneSvc.On("ListByApplicationTemplateVersionID", txtest.CtxWithDBMatcher(), appTemplateVersionID).Return(fixTombstones(), nil).Once()
//	return tombstoneSvc
//}
//
//successfulTombstoneUpdateForStaticDoc := func() *automockproc.TombstoneService {
//	tombstoneSvc := &automockproc.TombstoneService{}
//	tombstoneSvc.On("ListByApplicationTemplateVersionID", txtest.CtxWithDBMatcher(), appTemplateVersionID).Return(fixTombstones(), nil).Once()
//	tombstoneSvc.On("Update", txtest.CtxWithDBMatcher(), resource.ApplicationTemplateVersion, tombstoneID, *sanitizedStaticDoc.Tombstones[0]).Return(nil).Once()
//	tombstoneSvc.On("ListByApplicationTemplateVersionID", txtest.CtxWithDBMatcher(), appTemplateVersionID).Return(fixTombstones(), nil).Once()
//	return tombstoneSvc
//}

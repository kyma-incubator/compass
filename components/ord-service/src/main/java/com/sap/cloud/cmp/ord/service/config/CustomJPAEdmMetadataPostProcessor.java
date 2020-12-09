package com.sap.cloud.cmp.ord.service.config;

import com.sap.olingo.jpa.metadata.api.JPAEdmMetadataPostProcessor;
import com.sap.olingo.jpa.metadata.core.edm.mapper.exception.ODataJPAModelException;
import com.sap.olingo.jpa.metadata.core.edm.mapper.extention.IntermediateEntityTypeAccess;
import com.sap.olingo.jpa.metadata.core.edm.mapper.extention.IntermediateNavigationPropertyAccess;
import com.sap.olingo.jpa.metadata.core.edm.mapper.extention.IntermediatePropertyAccess;
import com.sap.olingo.jpa.metadata.core.edm.mapper.extention.IntermediateReferenceList;

public class CustomJPAEdmMetadataPostProcessor extends JPAEdmMetadataPostProcessor {

    public static String firstToLower(String substring) {
        return Character.toLowerCase(substring.charAt(0)) + substring.substring(1);
    }

    @Override
    public void processEntityType(IntermediateEntityTypeAccess property) {
        // EMPTY BODY
    }

    @Override
    public void processNavigationProperty(IntermediateNavigationPropertyAccess property, String s) {
        property.setExternalName(firstToLower(property.getInternalName()));
    }

    @Override
    public void processProperty(IntermediatePropertyAccess property, String jpaManagedTypeClassName) {
        property.setExternalName(firstToLower(property.getInternalName()));
    }

    @Override
    public void provideReferences(IntermediateReferenceList property) throws ODataJPAModelException {
        // EMPTY BODY
    }
}

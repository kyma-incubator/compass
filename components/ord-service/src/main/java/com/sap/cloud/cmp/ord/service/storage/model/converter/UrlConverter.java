package com.sap.cloud.cmp.ord.service.storage.model.converter;

import org.springframework.beans.factory.annotation.Value;

import javax.persistence.AttributeConverter;
import javax.persistence.Converter;

@Converter
public class UrlConverter implements AttributeConverter<String, String> {

    // @Value("${odata.jpa.request_mapping_path}") TODO: Use value from application.yaml
    private static String requestMappingPath = "open-resource-discovery";

    @Override
    public String convertToDatabaseColumn(String s) {
        return null; // ORD Service is read only
    }

    @Override
    public String convertToEntityAttribute(String s) {
        return "/" + requestMappingPath + s;
    }
}

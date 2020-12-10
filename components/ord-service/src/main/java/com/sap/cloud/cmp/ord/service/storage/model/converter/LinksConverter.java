package com.sap.cloud.cmp.ord.service.storage.model.converter;

import com.fasterxml.jackson.core.type.TypeReference;
import com.fasterxml.jackson.databind.ObjectMapper;
import com.sap.cloud.cmp.ord.service.storage.model.Link;
import org.postgresql.util.PGobject;
import org.springframework.beans.factory.annotation.Autowired;

import javax.persistence.AttributeConverter;
import javax.persistence.Converter;
import java.io.IOException;
import java.util.Collection;

@Converter
public class LinksConverter implements AttributeConverter<Collection<Link>, PGobject> {

    @Autowired
    ObjectMapper mapper;

    @Override
    public PGobject convertToDatabaseColumn(Collection<Link> links) {
        return null; // ORD Service has read only access.
    }

    @Override
    public Collection<Link> convertToEntityAttribute(PGobject databaseValue) {
        if (databaseValue == null)
            return null;
        try {
            return mapper.readerFor(new TypeReference<Collection<Link>>() {}).readValue(databaseValue.getValue());
        } catch (IOException e) {
            throw new IllegalArgumentException("Unable to deserialize to json field ", e);
        }
    }
}

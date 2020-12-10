package com.sap.cloud.cmp.ord.service.storage.model;

import org.eclipse.persistence.annotations.Convert;
import org.eclipse.persistence.annotations.TypeConverter;

import javax.persistence.*;
import java.util.UUID;

/**
 * Once we introduce ORD Aggregator APIDefinitions should be fetched as Specifications.
 * Therefore Compass should support multiple specs per API.
 * In order to achive that we will create a new specifications table which will unify api/event specifications.
 */
@Entity(name = "apiSpecification")
@Table(name = "api_definitions")
public class APISpecificationEntity {
    @javax.persistence.Id
    @Column(name = "id")
    @Convert("uuidConverter")
    @TypeConverter(name = "uuidConverter", dataType = Object.class, objectType = UUID.class)
    private UUID apiDefinitionId;

    @Column(name = "spec_data", length = Integer.MAX_VALUE)
    private String specData;

    @Column(name = "spec_format")
    private String specFormat;

    @Column(name = "spec_type")
    private String specType;

    @ManyToOne(fetch = FetchType.LAZY)
    @JoinColumn(name = "id", insertable = false, updatable = false)
    private APIEntity api;
}

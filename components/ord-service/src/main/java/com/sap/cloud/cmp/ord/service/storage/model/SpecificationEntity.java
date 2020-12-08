package com.sap.cloud.cmp.ord.service.storage.model;

import org.eclipse.persistence.annotations.Convert;
import org.eclipse.persistence.annotations.TypeConverter;

import javax.persistence.*;
import java.util.UUID;

@Entity(name = "Specification")
@Table(name="specifications")
public class SpecificationEntity {
    @javax.persistence.Id
    @Column(name = "id")
    @Convert("uuidConverter")
    @TypeConverter(name = "uuidConverter", dataType = Object.class, objectType = UUID.class)
    private UUID Id;

    @Column(name = "api_def_id")
    @Convert("uuidConverter")
    @TypeConverter(name = "uuidConverter", dataType = Object.class, objectType = UUID.class)
    private UUID apiDefId;

    @Column(name = "event_def_id")
    @Convert("uuidConverter")
    @TypeConverter(name = "uuidConverter", dataType = Object.class, objectType = UUID.class)
    private UUID eventDefId;

    @Column(name = "spec_data", length = Integer.MAX_VALUE)
    private String specData;

    @Column(name = "spec_format")
    private String specFormat;

    @Column(name = "spec_type")
    private String specType;

    @Column(name = "custom_type")
    private String customType;

    @ManyToOne(optional = true, fetch = FetchType.LAZY)
    @JoinColumn(name = "api_def_id", insertable = false, updatable = false)
    private APIEntity api;

    @ManyToOne(optional = true, fetch = FetchType.LAZY)
    @JoinColumn(name = "event_def_id", insertable = false, updatable = false)
    private EventEntity event;
}
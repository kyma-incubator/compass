package com.sap.cloud.cmp.ord.service.storage.model;

import org.eclipse.persistence.annotations.Convert;
import org.eclipse.persistence.annotations.TypeConverter;

import javax.persistence.*;
import javax.validation.constraints.NotNull;
import java.util.*;

@Entity(name = "package")
@Table(name = "packages")
public class PackageEntity {
    @javax.persistence.Id
    @Column(name = "id")
    @Convert("uuidConverter")
    @TypeConverter(name = "uuidConverter", dataType = Object.class, objectType = UUID.class)
    private UUID Id;

    @Column(name = "ord_id", length = 256)
    @NotNull
    private String ordId;

    @Column(name = "name", length = 256)
    @NotNull
    private String title;

    @Column(name = "short_description", length = 255)
    private String shortDescription;

    @Column(name = "description", length = Integer.MAX_VALUE)
    private String description;

    @Column(name = "version")
    private String version;

    @Column(name = "links", length = Integer.MAX_VALUE)
    private String links;

    @Column(name = "terms_of_service", length = 512)
    private String termsOfService;

    @Column(name = "licence_type", length = 256)
    private String licenceType;

    @Column(name = "licence", length = 512)
    private String licence;

    @Column(name = "provider", length = Integer.MAX_VALUE)
    private String provider;

    @Column(name = "tags", length = Integer.MAX_VALUE)
    private String tags;

    @Column(name = "actions", length = Integer.MAX_VALUE)
    private String actions;

    @Column(name = "extensions", length = Integer.MAX_VALUE)
    private String extensions;

    @OneToMany(mappedBy = "packageEntity", fetch = FetchType.LAZY)
    private Set<APIEntity> apis;

    @OneToMany(mappedBy = "packageEntity", fetch = FetchType.LAZY)
    private Set<EventEntity> events;
}
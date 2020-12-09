package com.sap.cloud.cmp.ord.service.storage.model;

import org.eclipse.persistence.annotations.Convert;
import org.eclipse.persistence.annotations.TypeConverter;

import javax.persistence.*;
import javax.validation.constraints.NotNull;
import java.util.UUID;

@Entity(name = "api")
@Table(name="api_definitions")
public class APIEntity {
    @javax.persistence.Id
    @Column(name = "id")
    @Convert("uuidConverter")
    @TypeConverter(name = "uuidConverter", dataType = Object.class, objectType = UUID.class)
    private UUID Id;

    @Column(name = "ord_id", length = 256)
    private String ordId;

    @Column(name = "name", length = 256)
    @NotNull
    private String title;

    @Column(name = "short_description", length = 255)
    private String shortDescription;

    @Column(name = "description", length = Integer.MAX_VALUE)
    private String description;

    @Column(name = "documentation", length = 512)
    private String documentation;

    @Column(name = "version_value")
    private String version;

    @Column(name = "system_instance_aware")
    private boolean systemInstanceAware;

    @Column(name = "package_id")
    @Convert("uuidConverter")
    @TypeConverter(name = "uuidConverter", dataType = Object.class, objectType = UUID.class)
    @NotNull
    private UUID partOfPackage;

    @Column(name = "api_protocol")
    private String apiProtocol;

    @Column(name = "tags", length = Integer.MAX_VALUE)
    private String tags;

    @Column(name = "api_definitions", length = Integer.MAX_VALUE)
    private String apiDefinitions;

    @Column(name = "links", length = Integer.MAX_VALUE)
    private String links;

    @Column(name = "actions", length = Integer.MAX_VALUE)
    private String actions;

    @Column(name = "release_status")
    @NotNull
    private String releaseStatus;

    @Column(name = "changelog_entries", length = Integer.MAX_VALUE)
    private String changelogEntries;

    @Column(name = "target_url", length = 256)
    @NotNull
    private String entryPoint;

    @Column(name = "extensions", length = Integer.MAX_VALUE)
    private String extensions;

    @ManyToOne(fetch = FetchType.LAZY)
    @JoinColumn(name = "package_id", insertable = false, updatable = false)
    private PackageEntity packageEntity;
}
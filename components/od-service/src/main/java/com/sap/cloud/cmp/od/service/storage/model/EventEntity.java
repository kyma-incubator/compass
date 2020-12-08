package com.sap.cloud.cmp.od.service.storage.model;

import org.eclipse.persistence.annotations.Convert;
import org.eclipse.persistence.annotations.TypeConverter;

import javax.persistence.*;
import javax.validation.constraints.NotNull;
import java.util.Set;
import java.util.UUID;

@Entity(name = "Event")
@Table(name="event_api_definitions")
public class EventEntity {
    @javax.persistence.Id
    @Column(name = "id")
    @Convert("uuidConverter")
    @TypeConverter(name = "uuidConverter", dataType = Object.class, objectType = UUID.class)
    private UUID Id;

    @Column(name = "od_id", length = 256)
    @NotNull
    private String openDiscoveryId;

    @Column(name = "bundle_id")
    @Convert("uuidConverter")
    @TypeConverter(name = "uuidConverter", dataType = Object.class, objectType = UUID.class)
    @NotNull
    private UUID bundleId;

    @Column(name = "title", length = 256)
    @NotNull
    private String title;

    @Column(name = "short_description", length = 256)
    @NotNull
    private String shortDescription;

    @Column(name = "description", length = Integer.MAX_VALUE)
    private String description;

    @Column(name = "event_definitions", length = Integer.MAX_VALUE)
    @NotNull
    private String apiDefinitions;

    @Column(name = "version")
    @NotNull
    private String version;

    @Column(name = "documentation", length = 512)
    private String documentation;

    @Column(name = "changelog_entries", length = Integer.MAX_VALUE)
    private String changelogEntries;

    @Column(name = "logo", length = 512)
    private String logo;

    @Column(name = "image", length = 512)
    private String image;

    @Column(name = "url", length = 512)
    private String url;

    @Column(name = "release_status")
    @NotNull
    private String releaseStatus;

    @Column(name = "tags", length = Integer.MAX_VALUE)
    private String tags;

    @Column(name = "extensions", length = Integer.MAX_VALUE)
    private String extensions;

    @ManyToOne(optional = true, fetch = FetchType.LAZY)
    @JoinColumn(name = "bundle_id", insertable = false, updatable = false)
    private BundleEntity bundle;

    @OneToMany(mappedBy = "event", fetch = FetchType.LAZY)
    private Set<SpecificationEntity> specifications;
}
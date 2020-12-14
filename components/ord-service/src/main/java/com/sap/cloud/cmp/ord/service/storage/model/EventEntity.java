package com.sap.cloud.cmp.ord.service.storage.model;

import org.eclipse.persistence.annotations.Convert;
import org.eclipse.persistence.annotations.TypeConverter;

import javax.persistence.*;
import javax.validation.constraints.NotNull;
import java.util.List;
import java.util.Set;
import java.util.UUID;

@Entity(name = "event")
@Table(name = "event_api_definitions")
public class EventEntity {
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

    @Column(name = "version_value")
    private String version;

    @Column(name = "system_instance_aware")
    private boolean systemInstanceAware;

    @ElementCollection
    @CollectionTable(name="changelog_entries", joinColumns=@JoinColumn(name="event_definition_id"))
    private List<ChangelogEntry> changelogEntries;

    @Column(name = "package_id")
    @Convert("uuidConverter")
    @TypeConverter(name = "uuidConverter", dataType = Object.class, objectType = UUID.class)
    @NotNull
    private UUID partOfPackage;

    @ElementCollection
    @CollectionTable(name="links", joinColumns=@JoinColumn(name="event_definition_id"))
    private List<Link> links;

    @ElementCollection
    @CollectionTable(name = "tags", joinColumns = @JoinColumn(name = "event_definition_id"))
    private List<Tag> tags;

    @Column(name = "release_status")
    @NotNull
    private String releaseStatus;

    @ElementCollection
    @CollectionTable(name="event_resource_definitions", joinColumns=@JoinColumn(name="event_definition_id"))
    private List<EventDefinition> eventDefinitions;

    @Column(name = "extensions", length = Integer.MAX_VALUE)
    private String extensions;

    @OneToMany(mappedBy = "event", fetch = FetchType.LAZY)
    private Set<EventSpecificationEntity> specifications;

    @ManyToOne(fetch = FetchType.LAZY)
    @JoinColumn(name = "package_id", insertable = false, updatable = false)
    private PackageEntity packageEntity;
}
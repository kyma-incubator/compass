package com.sap.cloud.cmp.ord.service.storage.model;

import org.eclipse.persistence.annotations.Convert;
import org.eclipse.persistence.annotations.TypeConverter;

import javax.persistence.*;
import javax.validation.constraints.NotNull;
import java.util.List;
import java.util.Set;
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

    @ElementCollection
    @CollectionTable(name = "tags", joinColumns = @JoinColumn(name = "api_definition_id"))
    private List<ArrayElement> tags;

    @ElementCollection
    @CollectionTable(name = "countries", joinColumns = @JoinColumn(name = "api_definition_id"))
    private List<ArrayElement> countries;


    @ElementCollection
    @CollectionTable(name="api_resource_definitions", joinColumns=@JoinColumn(name="api_definition_id"))
    private List<APIDefinition> apiDefinitions;

    @ElementCollection
    @CollectionTable(name="links", joinColumns=@JoinColumn(name="api_definition_id"))
    private List<Link> links;

    @ElementCollection
    @CollectionTable(name="api_resource_links", joinColumns=@JoinColumn(name="api_definition_id"))
    private List<APIResourceLink> apiResourceLinks;

    @Column(name = "release_status")
    @NotNull
    private String releaseStatus;

    @Column(name = "sunset_date")
    private String sunsetDate;

    @Column(name = "successor")
    private String successor;

    @ElementCollection
    @CollectionTable(name="changelog_entries", joinColumns=@JoinColumn(name="api_definition_id"))
    private List<ChangelogEntry> changelogEntries;

    @Column(name = "target_url", length = 256)
    @NotNull
    private String entryPoint;

    @ElementCollection
    @CollectionTable(name = "ord_labels", joinColumns = @JoinColumn(name = "api_definition_id"))
    private List<Label> labels;

    @OneToMany(mappedBy = "api", fetch = FetchType.LAZY)
    private Set<APISpecificationEntity> specifications;

    @ManyToOne(fetch = FetchType.LAZY)
    @JoinColumn(name = "package_id", insertable = false, updatable = false)
    private PackageEntity packageEntity;
}
package com.sap.cloud.cmp.ord.service.storage.model;

import org.eclipse.persistence.annotations.Convert;
import org.eclipse.persistence.annotations.TypeConverter;

import javax.persistence.*;
import javax.validation.constraints.NotNull;
import java.util.*;

@Entity(name = "package")
@Table(name = "packages")
@SecondaryTable(name = "providers", pkJoinColumns = @PrimaryKeyJoinColumn(name = "package_id", referencedColumnName = "id"))
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

    @ElementCollection
    @CollectionTable(name = "links", joinColumns = @JoinColumn(name = "package_id"))
    private List<Link> links;

    @Column(name = "terms_of_service", length = 512)
    private String termsOfService;

    @Column(name = "licence_type", length = 256)
    private String licenceType;

    @Column(name = "licence", length = 512)
    private String licence;

    @Embedded
    @AttributeOverrides({
            @AttributeOverride(name = "name", column = @Column(name = "name", table = "providers")),
            @AttributeOverride(name = "department", column = @Column(name = "department", table = "providers")),
            @AttributeOverride(name = "extensions", column = @Column(name = "extensions", table = "providers")),
    })
    private Provider provider;

    @ElementCollection
    @CollectionTable(name = "tags", joinColumns = @JoinColumn(name = "package_id"))
    private List<Tag> tags;

    @Column(name = "actions", length = Integer.MAX_VALUE)
    private String actions;

    @Column(name = "extensions", length = Integer.MAX_VALUE)
    private String extensions;

    @OneToMany(mappedBy = "packageEntity", fetch = FetchType.LAZY)
    private Set<APIEntity> apis;

    @OneToMany(mappedBy = "packageEntity", fetch = FetchType.LAZY)
    private Set<EventEntity> events;
}
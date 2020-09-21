package com.sap.cloud.cmp.od.service.storage.model;

import org.eclipse.persistence.annotations.Convert;
import org.eclipse.persistence.annotations.TypeConverter;

import javax.persistence.*;
import javax.validation.constraints.NotNull;
import java.util.Set;
import java.util.UUID;

@Entity (name = "Bundle")
@Table(name="bundles")
public class BundleEntity {
    @javax.persistence.Id
    @Column(name = "id")
    @Convert("uuidConverter")
    @TypeConverter(name = "uuidConverter", dataType = Object.class, objectType = UUID.class)
    private UUID Id;

    @Column(name = "od_id", length = 256)
    @NotNull
    private String openDiscoveryId;

    @Column(name = "title", length = 256)
    @NotNull
    private String title;

    @Column(name = "short_description", length = 256)
    @NotNull
    private String shortDescription;

    @Column(name = "description", length = Integer.MAX_VALUE)
    private String description;

    @Column(name = "instance_auth_request_json_schema", length = Integer.MAX_VALUE)
    @NotNull
    private String instanceAuthRequestJSONSchema;

    @Column(name = "default_instance_auth", length = Integer.MAX_VALUE)
    @NotNull
    private String defaultInstanceAuth;

    @Column(name = "tags", length = Integer.MAX_VALUE)
    private String tags;

    @Column(name = "extensions", length = Integer.MAX_VALUE)
    private String extensions;

    @OneToMany(mappedBy = "bundle", fetch = FetchType.LAZY)
    private Set<APIEntity> apis;

    @OneToMany(mappedBy = "bundle", fetch = FetchType.LAZY)
    private Set<EventEntity> events;

    @ManyToMany
    @JoinTable(
            name = "package_bundles",
            joinColumns = @JoinColumn(name = "bundle_id"),
            inverseJoinColumns = @JoinColumn(name = "package_id"))
    Set<PackageEntity> packages;
}
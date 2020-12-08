package com.sap.cloud.cmp.ord.service.storage.model;

import org.eclipse.persistence.annotations.Convert;
import org.eclipse.persistence.annotations.TypeConverter;

import javax.persistence.*;
import javax.validation.constraints.NotNull;
import java.util.Set;
import java.util.UUID;

@Entity(name = "Package")
@Table(name="packages")
public class PackageEntity {
    @javax.persistence.Id
    @Column(name = "id")
    @Convert("uuidConverter")
    @TypeConverter(name = "uuidConverter", dataType = Object.class, objectType = UUID.class)
    private UUID Id;

    @Column(name = "ord_id", length = 256)
    @NotNull
    private String openDiscoveryId;

    @Column(name = "title", length = 256)
    @NotNull
    private String title;

    @Column(name = "short_description", length = 256)
    @NotNull
    private String shortDescription;

    @Column(name = "description", length = Integer.MAX_VALUE)
    @NotNull
    private String description;

    @Column(name = "version")
    @NotNull
    private String version;

    @Column(name = "licence", length = 512)
    private String licence;

    @Column(name = "licence_type", length = 256)
    private String licenceType;

    @Column(name = "terms_of_service", length = 256)
    private String termsOfService;

    @Column(name = "logo", length = 512)
    private String logo;

    @Column(name = "image", length = 512)
    private String image;

    @Column(name = "provider", length = Integer.MAX_VALUE)
    private String provider;

    @Column(name = "tags", length = Integer.MAX_VALUE)
    private String tags;

    @Column(name = "actions", length = Integer.MAX_VALUE)
    private String actions;

    @Column(name = "extensions", length = Integer.MAX_VALUE)
    private String extensions;

    @ManyToMany(mappedBy = "packages")
    Set<BundleEntity> bundles;
}
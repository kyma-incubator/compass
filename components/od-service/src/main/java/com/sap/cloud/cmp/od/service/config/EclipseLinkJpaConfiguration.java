package com.sap.cloud.cmp.od.service.config;

import org.eclipse.persistence.config.PersistenceUnitProperties;
import org.eclipse.persistence.logging.SessionLog;
import org.reflections.Reflections;
import org.reflections.scanners.SubTypesScanner;
import org.reflections.scanners.TypeAnnotationsScanner;
import org.springframework.beans.factory.ObjectProvider;
import org.springframework.beans.factory.annotation.Value;
import org.springframework.boot.autoconfigure.orm.jpa.JpaBaseConfiguration;
import org.springframework.boot.autoconfigure.orm.jpa.JpaProperties;
import org.springframework.boot.autoconfigure.transaction.TransactionManagerCustomizers;
import org.springframework.boot.orm.jpa.EntityManagerFactoryBuilder;
import org.springframework.context.annotation.Bean;
import org.springframework.context.annotation.Configuration;
import org.springframework.orm.jpa.LocalContainerEntityManagerFactoryBean;
import org.springframework.orm.jpa.vendor.AbstractJpaVendorAdapter;
import org.springframework.orm.jpa.vendor.EclipseLinkJpaVendorAdapter;
import org.springframework.transaction.jta.JtaTransactionManager;

import javax.persistence.Entity;
import javax.sql.DataSource;
import java.util.HashMap;
import java.util.Map;
import java.util.Set;

@Configuration
public class EclipseLinkJpaConfiguration extends JpaBaseConfiguration {

    @Value("${odata.jpa.punit_name}")
    private String punit;

    @Value("${odata.jpa.root_packages}")
    private String rootPackages;

    private static final String ENTITIES_SUB_PACKAGE = ".storage.model";

    protected EclipseLinkJpaConfiguration(DataSource dataSource, JpaProperties properties,
                                          ObjectProvider<JtaTransactionManager> jtaTransactionManager,
                                          ObjectProvider<TransactionManagerCustomizers> transactionManagerCustomizers) {
        super(dataSource, properties, jtaTransactionManager, transactionManagerCustomizers);
    }

    @Override
    protected AbstractJpaVendorAdapter createJpaVendorAdapter() {
        return new EclipseLinkJpaVendorAdapter();
    }

    @Override
    protected Map<String, Object> getVendorProperties() {
        // https://stackoverflow.com/questions/10769051/eclipselinkjpavendoradapter-instead-of-hibernatejpavendoradapter-issue
        HashMap<String, Object> map = new HashMap<>();
        map.put(PersistenceUnitProperties.WEAVING, "false");
        map.put(PersistenceUnitProperties.DDL_GENERATION, "none");
        map.put(PersistenceUnitProperties.LOGGING_LEVEL, SessionLog.FINE_LABEL);
        map.put(PersistenceUnitProperties.TRANSACTION_TYPE, "RESOURCE_LOCAL");
        return map;
    }

    @Bean
    public LocalContainerEntityManagerFactoryBean customerEntityManagerFactory(
            final EntityManagerFactoryBuilder builder, final DataSource ds) {
        Set<Class<?>> entityClasses = getEntityClasses();

        return builder
                .dataSource(ds)
                .packages(entityClasses.toArray(new Class<?>[entityClasses.size()]))
                .persistenceUnit(punit)
                .properties(getVendorProperties())
                .jta(false)
                .build();
    }

    private Set<Class<?>> getEntityClasses() {
        Reflections reflections = new Reflections(rootPackages + ENTITIES_SUB_PACKAGE,
                new TypeAnnotationsScanner(), new SubTypesScanner());
        return reflections.getTypesAnnotatedWith(Entity.class);
    }
}

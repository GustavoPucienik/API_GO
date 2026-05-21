-- ── SAP sync tables ───────────────────────────────────────────────────────────

CREATE TABLE IF NOT EXISTS sap_client (
    sap_client_code  VARCHAR(50)  NOT NULL,
    name             VARCHAR(255) NOT NULL,
    alias_name       VARCHAR(255) NULL,
    sap_updated_at   DATETIME     NULL,
    created_at       DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at       DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    deleted_at       DATETIME     NULL,
    PRIMARY KEY (sap_client_code)
);

CREATE TABLE IF NOT EXISTS sap_client_contact (
    sap_client_contact_id INT          NOT NULL AUTO_INCREMENT,
    sap_client_code       VARCHAR(50)  NOT NULL,
    name                  VARCHAR(255) NOT NULL,
    contact_name          VARCHAR(255) NULL,
    phone                 VARCHAR(50)  NULL,
    mobile                VARCHAR(50)  NULL,
    email                 VARCHAR(255) NULL,
    created_at            DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at            DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    deleted_at            DATETIME     NULL,
    PRIMARY KEY (sap_client_contact_id),
    UNIQUE KEY uq_sap_contact (sap_client_code, name),
    FOREIGN KEY (sap_client_code) REFERENCES sap_client(sap_client_code)
);

CREATE TABLE IF NOT EXISTS sap_client_address (
    sap_client_address_id INT          NOT NULL AUTO_INCREMENT,
    sap_client_code       VARCHAR(50)  NOT NULL,
    address_name          VARCHAR(100) NOT NULL,
    address_type          VARCHAR(50)  NULL,
    type_of_address       VARCHAR(100) NULL,
    street                VARCHAR(255) NULL,
    street_no             VARCHAR(50)  NULL,
    block                 VARCHAR(100) NULL,
    city                  VARCHAR(100) NULL,
    county                VARCHAR(100) NULL,
    state                 VARCHAR(100) NULL,
    country               VARCHAR(100) NULL,
    zip_code              VARCHAR(20)  NULL,
    building_floor_room   VARCHAR(255) NULL,
    created_at            DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at            DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    deleted_at            DATETIME     NULL,
    PRIMARY KEY (sap_client_address_id),
    UNIQUE KEY uq_sap_address (sap_client_code, address_name, address_type),
    FOREIGN KEY (sap_client_code) REFERENCES sap_client(sap_client_code)
);

CREATE TABLE IF NOT EXISTS sap_item (
    sap_item_code  VARCHAR(50)  NOT NULL,
    name           VARCHAR(255) NOT NULL,
    sap_updated_at DATETIME     NULL,
    created_at     DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at     DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    deleted_at     DATETIME     NULL,
    PRIMARY KEY (sap_item_code)
);

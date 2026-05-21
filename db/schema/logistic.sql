-- ── Logistic module ───────────────────────────────────────────────────────────

CREATE TABLE IF NOT EXISTS logistic_order_status (
    logistic_order_status_id INT          NOT NULL AUTO_INCREMENT,
    name                     VARCHAR(100) NOT NULL UNIQUE,
    background_color         VARCHAR(20)  NOT NULL,
    border_color             VARCHAR(20)  NOT NULL,
    text_color               VARCHAR(20)  NOT NULL,
    sort_index               INT          NOT NULL DEFAULT 0,
    is_closed                BOOLEAN      NOT NULL DEFAULT FALSE,
    is_pending_approval      BOOLEAN      NOT NULL DEFAULT FALSE,
    is_rh_step               BOOLEAN      NOT NULL DEFAULT FALSE,
    PRIMARY KEY (logistic_order_status_id)
);

CREATE TABLE IF NOT EXISTS logistic_order_priorities (
    logistic_order_priority_id INT          NOT NULL AUTO_INCREMENT,
    name                       VARCHAR(100) NOT NULL UNIQUE,
    background_color           VARCHAR(20)  NOT NULL,
    border_color               VARCHAR(20)  NOT NULL,
    text_color                 VARCHAR(20)  NOT NULL,
    sort_index                 INT          NOT NULL DEFAULT 0,
    PRIMARY KEY (logistic_order_priority_id)
);

CREATE TABLE IF NOT EXISTS logistic_order_reason (
    logistic_order_reason_id INT          NOT NULL AUTO_INCREMENT,
    name                     VARCHAR(100) NOT NULL UNIQUE,
    initial_status_id        INT          NOT NULL,
    created_at               DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at               DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    deleted_at               DATETIME     NULL,
    PRIMARY KEY (logistic_order_reason_id),
    FOREIGN KEY (initial_status_id) REFERENCES logistic_order_status(logistic_order_status_id)
);

CREATE TABLE IF NOT EXISTS logistic_order_status_transition (
    id               INT NOT NULL AUTO_INCREMENT,
    from_status_id   INT NULL,
    to_status_id     INT NOT NULL,
    access_level_id  INT NOT NULL,
    PRIMARY KEY (id),
    UNIQUE KEY uq_logistic_transition (from_status_id, to_status_id, access_level_id),
    FOREIGN KEY (from_status_id)  REFERENCES logistic_order_status(logistic_order_status_id),
    FOREIGN KEY (to_status_id)    REFERENCES logistic_order_status(logistic_order_status_id),
    FOREIGN KEY (access_level_id) REFERENCES access_level(access_level_id)
);

CREATE TABLE IF NOT EXISTS logistic_operator (
    logistic_operator_id INT                             NOT NULL AUTO_INCREMENT,
    name                 VARCHAR(255)                    NOT NULL,
    type                 ENUM('driver','checker','both') NOT NULL DEFAULT 'both',
    PRIMARY KEY (logistic_operator_id)
);

CREATE TABLE IF NOT EXISTS logistic_order (
    logistic_order_id       INT          NOT NULL AUTO_INCREMENT,
    requester_id            INT          NOT NULL,
    requester_name          VARCHAR(255) NOT NULL,
    assigned_to_id          INT          NULL,
    assigned_to_name        VARCHAR(255) NULL,
    driver_name             VARCHAR(255) NULL,
    checker_name            VARCHAR(255) NULL,
    checker_signature       TEXT         NULL,
    salesperson_name        VARCHAR(255) NULL,
    client_code             VARCHAR(50)  NOT NULL,
    client_name             VARCHAR(255) NOT NULL,
    client_alias_name       VARCHAR(255) NOT NULL,
    contact_type            VARCHAR(100) NULL,
    contact_name            VARCHAR(255) NULL,
    contact_phone           VARCHAR(50)  NULL,
    contact_mobile          VARCHAR(50)  NULL,
    contact_email           VARCHAR(255) NULL,
    address_name            VARCHAR(100) NULL,
    address_type            VARCHAR(50)  NULL,
    type_of_address         VARCHAR(100) NULL,
    street                  VARCHAR(255) NULL,
    street_no               VARCHAR(50)  NULL,
    block                   VARCHAR(100) NULL,
    city                    VARCHAR(100) NULL,
    county                  VARCHAR(100) NULL,
    state                   VARCHAR(100) NULL,
    country                 VARCHAR(100) NULL,
    zip_code                VARCHAR(20)  NULL,
    building_floor_room     VARCHAR(255) NULL,
    status_id               INT          NOT NULL,
    priority_id             INT          NOT NULL,
    reason_id               INT          NOT NULL,
    schedule_date           DATETIME     NULL,
    description             TEXT         NOT NULL,
    has_salary_discount     BOOLEAN      NULL DEFAULT NULL,
    salary_discount_reason  TEXT         NULL,
    captured_at             DATETIME     NULL,
    resolution              TEXT         NULL,
    closed_at               DATETIME     NULL,
    created_at              DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at              DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    deleted_at              DATETIME     NULL,
    PRIMARY KEY (logistic_order_id),
    FOREIGN KEY (status_id)   REFERENCES logistic_order_status(logistic_order_status_id),
    FOREIGN KEY (priority_id) REFERENCES logistic_order_priorities(logistic_order_priority_id),
    FOREIGN KEY (reason_id)   REFERENCES logistic_order_reason(logistic_order_reason_id)
);

CREATE TABLE IF NOT EXISTS logistic_order_items (
    logistic_order_item_id INT          NOT NULL AUTO_INCREMENT,
    logistic_order_id      INT          NOT NULL,
    name                   VARCHAR(255) NOT NULL,
    item_code              VARCHAR(50)  NOT NULL,
    quantity               INT          NOT NULL DEFAULT 1,
    created_at             DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at             DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    deleted_at             DATETIME     NULL,
    PRIMARY KEY (logistic_order_item_id),
    FOREIGN KEY (logistic_order_id) REFERENCES logistic_order(logistic_order_id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS logistic_order_attachment (
    logistic_order_attachment_id INT          NOT NULL AUTO_INCREMENT,
    logistic_order_id            INT          NOT NULL,
    file_name                    VARCHAR(255) NOT NULL,
    file_path                    VARCHAR(500) NOT NULL,
    file_size                    INT          NOT NULL,
    created_at                   DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at                   DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    deleted_at                   DATETIME     NULL,
    PRIMARY KEY (logistic_order_attachment_id),
    FOREIGN KEY (logistic_order_id) REFERENCES logistic_order(logistic_order_id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS logistic_config (
    logistic_config_id               INT NOT NULL AUTO_INCREMENT,
    salary_discount_access_level_id  INT NULL,
    PRIMARY KEY (logistic_config_id),
    FOREIGN KEY (salary_discount_access_level_id) REFERENCES access_level(access_level_id) ON DELETE SET NULL
);

CREATE TABLE IF NOT EXISTS logistic_config_assignable_access_level (
    logistic_config_id INT NOT NULL,
    access_level_id    INT NOT NULL,
    PRIMARY KEY (logistic_config_id, access_level_id),
    FOREIGN KEY (logistic_config_id) REFERENCES logistic_config(logistic_config_id) ON DELETE CASCADE,
    FOREIGN KEY (access_level_id)    REFERENCES access_level(access_level_id) ON DELETE CASCADE
);

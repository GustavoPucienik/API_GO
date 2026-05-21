-- ── Return module ─────────────────────────────────────────────────────────────

CREATE TABLE IF NOT EXISTS return_order_status (
    return_order_status_id INT          NOT NULL AUTO_INCREMENT,
    name                   VARCHAR(100) NOT NULL UNIQUE,
    background_color       VARCHAR(20)  NOT NULL,
    border_color           VARCHAR(20)  NOT NULL,
    text_color             VARCHAR(20)  NOT NULL,
    sort_index             INT          NOT NULL DEFAULT 0,
    is_closed              BOOLEAN      NOT NULL DEFAULT FALSE,
    is_default             BOOLEAN      NOT NULL DEFAULT FALSE,
    is_pending_approval    BOOLEAN      NOT NULL DEFAULT FALSE,
    PRIMARY KEY (return_order_status_id)
);

CREATE TABLE IF NOT EXISTS return_order_priorities (
    return_order_priority_id INT          NOT NULL AUTO_INCREMENT,
    name                     VARCHAR(100) NOT NULL UNIQUE,
    background_color         VARCHAR(20)  NOT NULL,
    border_color             VARCHAR(20)  NOT NULL,
    text_color               VARCHAR(20)  NOT NULL,
    sort_index               INT          NOT NULL DEFAULT 0,
    PRIMARY KEY (return_order_priority_id)
);

CREATE TABLE IF NOT EXISTS return_order_reason (
    return_order_reason_id INT          NOT NULL AUTO_INCREMENT,
    name                   VARCHAR(100) NOT NULL UNIQUE,
    description            TEXT         NULL,
    initial_status_id      INT          NOT NULL,
    created_at             DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at             DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    deleted_at             DATETIME     NULL,
    PRIMARY KEY (return_order_reason_id),
    FOREIGN KEY (initial_status_id) REFERENCES return_order_status(return_order_status_id)
);

CREATE TABLE IF NOT EXISTS return_order_credit_form (
    return_order_credit_form_id INT          NOT NULL AUTO_INCREMENT,
    name                        VARCHAR(100) NOT NULL UNIQUE,
    created_at                  DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at                  DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    deleted_at                  DATETIME     NULL,
    PRIMARY KEY (return_order_credit_form_id)
);

CREATE TABLE IF NOT EXISTS return_order_status_transition (
    id               INT NOT NULL AUTO_INCREMENT,
    from_status_id   INT NULL,
    to_status_id     INT NOT NULL,
    access_level_id  INT NOT NULL,
    PRIMARY KEY (id),
    UNIQUE KEY uq_return_transition (from_status_id, to_status_id, access_level_id),
    FOREIGN KEY (from_status_id)  REFERENCES return_order_status(return_order_status_id),
    FOREIGN KEY (to_status_id)    REFERENCES return_order_status(return_order_status_id),
    FOREIGN KEY (access_level_id) REFERENCES access_level(access_level_id)
);

CREATE TABLE IF NOT EXISTS return_order (
    return_order_id          INT          NOT NULL AUTO_INCREMENT,
    requester_id             INT          NOT NULL,
    requester_name           VARCHAR(255) NOT NULL,
    assigned_to_id           INT          NULL,
    assigned_to_name         VARCHAR(255) NULL,
    salesperson_name         VARCHAR(255) NULL,
    client_code              VARCHAR(50)  NOT NULL,
    client_name              VARCHAR(255) NOT NULL,
    client_alias_name        VARCHAR(255) NOT NULL,
    contact_type             VARCHAR(100) NULL,
    contact_name             VARCHAR(255) NULL,
    contact_phone            VARCHAR(50)  NULL,
    contact_mobile           VARCHAR(50)  NULL,
    contact_email            VARCHAR(255) NULL,
    address_name             VARCHAR(100) NULL,
    address_type             VARCHAR(50)  NULL,
    type_of_address          VARCHAR(100) NULL,
    street                   VARCHAR(255) NULL,
    street_no                VARCHAR(50)  NULL,
    block                    VARCHAR(100) NULL,
    city                     VARCHAR(100) NULL,
    county                   VARCHAR(100) NULL,
    state                    VARCHAR(100) NULL,
    country                  VARCHAR(100) NULL,
    zip_code                 VARCHAR(20)  NULL,
    building_floor_room      VARCHAR(255) NULL,
    status_id                INT          NOT NULL,
    priority_id              INT          NOT NULL,
    reason_id                INT          NOT NULL,
    credit_form_id           INT          NULL,
    pix_key_type             VARCHAR(50)  NULL,
    pix_key_owner            VARCHAR(255) NULL,
    pix_key                  VARCHAR(255) NULL,
    transfer_bank            VARCHAR(100) NULL,
    transfer_agency          VARCHAR(50)  NULL,
    transfer_account         VARCHAR(50)  NULL,
    nfe                      VARCHAR(100) NULL,
    client_sent_return_note  BOOLEAN      NOT NULL DEFAULT FALSE,
    internal_return          BOOLEAN      NOT NULL DEFAULT FALSE,
    partial_return           BOOLEAN      NOT NULL DEFAULT FALSE,
    schedule_date            DATETIME     NULL,
    description              TEXT         NOT NULL,
    captured_at              DATETIME     NULL,
    resolution               TEXT         NULL,
    closed_at                DATETIME     NULL,
    created_at               DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at               DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    deleted_at               DATETIME     NULL,
    PRIMARY KEY (return_order_id),
    FOREIGN KEY (status_id)      REFERENCES return_order_status(return_order_status_id),
    FOREIGN KEY (priority_id)    REFERENCES return_order_priorities(return_order_priority_id),
    FOREIGN KEY (reason_id)      REFERENCES return_order_reason(return_order_reason_id),
    FOREIGN KEY (credit_form_id) REFERENCES return_order_credit_form(return_order_credit_form_id) ON DELETE SET NULL
);

CREATE TABLE IF NOT EXISTS return_order_items (
    return_order_item_id INT          NOT NULL AUTO_INCREMENT,
    return_order_id      INT          NOT NULL,
    name                 VARCHAR(255) NOT NULL,
    item_code            VARCHAR(50)  NOT NULL,
    quantity             INT          NOT NULL DEFAULT 1,
    created_at           DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at           DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    deleted_at           DATETIME     NULL,
    PRIMARY KEY (return_order_item_id),
    FOREIGN KEY (return_order_id) REFERENCES return_order(return_order_id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS return_order_attachment (
    return_order_attachment_id INT          NOT NULL AUTO_INCREMENT,
    return_order_id            INT          NOT NULL,
    file_name                  VARCHAR(255) NOT NULL,
    file_path                  VARCHAR(500) NOT NULL,
    file_size                  INT          NOT NULL,
    created_at                 DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at                 DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    deleted_at                 DATETIME     NULL,
    PRIMARY KEY (return_order_attachment_id),
    FOREIGN KEY (return_order_id) REFERENCES return_order(return_order_id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS return_config (
    return_config_id INT NOT NULL AUTO_INCREMENT,
    PRIMARY KEY (return_config_id)
);

CREATE TABLE IF NOT EXISTS return_config_assignable_access_level (
    return_config_id INT NOT NULL,
    access_level_id  INT NOT NULL,
    PRIMARY KEY (return_config_id, access_level_id),
    FOREIGN KEY (return_config_id) REFERENCES return_config(return_config_id) ON DELETE CASCADE,
    FOREIGN KEY (access_level_id)  REFERENCES access_level(access_level_id) ON DELETE CASCADE
);

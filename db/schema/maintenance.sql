-- ── Lookups ──────────────────────────────────────────────────────────────────

CREATE TABLE IF NOT EXISTS maintenance_order_status (
    maintenance_order_status_id INT          NOT NULL AUTO_INCREMENT,
    name                        VARCHAR(100) NOT NULL UNIQUE,
    background_color            VARCHAR(20)  NOT NULL,
    border_color                VARCHAR(20)  NOT NULL,
    text_color                  VARCHAR(20)  NOT NULL,
    sort_index                  INT          NOT NULL DEFAULT 0,
    is_closed                   TINYINT(1)   NOT NULL DEFAULT 0,
    is_default                  TINYINT(1)   NOT NULL DEFAULT 0,
    PRIMARY KEY (maintenance_order_status_id)
);

CREATE TABLE IF NOT EXISTS maintenance_order_priorities (
    maintenance_order_priorities_id INT          NOT NULL AUTO_INCREMENT,
    name                            VARCHAR(100) NOT NULL UNIQUE,
    background_color                VARCHAR(20)  NOT NULL,
    border_color                    VARCHAR(20)  NOT NULL,
    text_color                      VARCHAR(20)  NOT NULL,
    sort_index                      INT          NOT NULL DEFAULT 0,
    PRIMARY KEY (maintenance_order_priorities_id)
);

CREATE TABLE IF NOT EXISTS maintenance_order_reason (
    maintenance_order_reason_id INT          NOT NULL AUTO_INCREMENT,
    name                        VARCHAR(100) NOT NULL UNIQUE,
    created_at                  DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at                  DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    deleted_at                  DATETIME     NULL,
    PRIMARY KEY (maintenance_order_reason_id)
);

CREATE TABLE IF NOT EXISTS maintenance_order_note_template (
    maintenance_order_note_template_id INT          NOT NULL AUTO_INCREMENT,
    name                               VARCHAR(255) NOT NULL UNIQUE,
    description                        TEXT         NOT NULL,
    created_at                         DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at                         DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    deleted_at                         DATETIME     NULL,
    PRIMARY KEY (maintenance_order_note_template_id)
);

CREATE TABLE IF NOT EXISTS maintenance_order_time_entry_type (
    maintenance_order_time_entry_type_id INT          NOT NULL AUTO_INCREMENT,
    name                                 VARCHAR(100) NOT NULL UNIQUE,
    color                                VARCHAR(20)  NULL,
    PRIMARY KEY (maintenance_order_time_entry_type_id)
);

-- ── Main order ────────────────────────────────────────────────────────────────

CREATE TABLE IF NOT EXISTS maintenance_order (
    maintenance_order_id   INT          NOT NULL AUTO_INCREMENT,
    requester_id           INT          NOT NULL,
    requester_name         VARCHAR(255) NOT NULL,
    salesperson_name       VARCHAR(255) NULL,
    client_code            VARCHAR(50)  NOT NULL,
    client_name            VARCHAR(255) NOT NULL,
    client_alias_name      VARCHAR(255) NOT NULL,
    contact_type           VARCHAR(100) NULL,
    contact_name           VARCHAR(255) NULL,
    contact_phone          VARCHAR(50)  NULL,
    contact_mobile         VARCHAR(50)  NULL,
    contact_email          VARCHAR(255) NULL,
    address_name           VARCHAR(100) NULL,
    address_type           VARCHAR(50)  NULL,
    type_of_address        VARCHAR(100) NULL,
    street                 VARCHAR(255) NULL,
    street_no              VARCHAR(50)  NULL,
    block                  VARCHAR(100) NULL,
    city                   VARCHAR(100) NULL,
    county                 VARCHAR(100) NULL,
    state                  VARCHAR(100) NULL,
    country                VARCHAR(100) NULL,
    zip_code               VARCHAR(20)  NULL,
    building_floor_room    VARCHAR(255) NULL,
    status_id              INT          NOT NULL,
    priority_id            INT          NOT NULL,
    reason_id              INT          NOT NULL,
    description            TEXT         NOT NULL,
    captured_at            DATETIME     NULL,
    resolution             TEXT         NULL,
    motivo_resolution      VARCHAR(255) NULL,
    client_signature       LONGTEXT     NULL,
    closed_at              DATETIME     NULL,
    created_at             DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at             DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    deleted_at             DATETIME     NULL,
    PRIMARY KEY (maintenance_order_id),
    CONSTRAINT fk_mo_status   FOREIGN KEY (status_id)   REFERENCES maintenance_order_status      (maintenance_order_status_id),
    CONSTRAINT fk_mo_priority FOREIGN KEY (priority_id) REFERENCES maintenance_order_priorities  (maintenance_order_priorities_id),
    CONSTRAINT fk_mo_reason   FOREIGN KEY (reason_id)   REFERENCES maintenance_order_reason      (maintenance_order_reason_id)
);

-- ── Items ─────────────────────────────────────────────────────────────────────

CREATE TABLE IF NOT EXISTS maintenance_order_items (
    maintenance_order_item_id INT          NOT NULL AUTO_INCREMENT,
    maintenance_order_id      INT          NOT NULL,
    name                      VARCHAR(255) NOT NULL,
    item_code                 VARCHAR(50)  NOT NULL,
    quantity                  INT          NOT NULL DEFAULT 1,
    created_at                DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at                DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    deleted_at                DATETIME     NULL,
    PRIMARY KEY (maintenance_order_item_id),
    CONSTRAINT fk_moi_order FOREIGN KEY (maintenance_order_id) REFERENCES maintenance_order (maintenance_order_id)
);

CREATE TABLE IF NOT EXISTS maintenance_order_resolution_items (
    maintenance_order_resolution_item_id INT          NOT NULL AUTO_INCREMENT,
    maintenance_order_id                 INT          NOT NULL,
    name                                 VARCHAR(255) NOT NULL,
    item_code                            VARCHAR(50)  NOT NULL,
    quantity                             INT          NOT NULL DEFAULT 1,
    created_at                           DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at                           DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    deleted_at                           DATETIME     NULL,
    PRIMARY KEY (maintenance_order_resolution_item_id),
    CONSTRAINT fk_mori_order FOREIGN KEY (maintenance_order_id) REFERENCES maintenance_order (maintenance_order_id)
);

-- ── Time entries ──────────────────────────────────────────────────────────────

CREATE TABLE IF NOT EXISTS maintenance_order_time_entry (
    maintenance_order_time_entry_id INT       NOT NULL AUTO_INCREMENT,
    maintenance_order_id            INT       NOT NULL,
    type_id                         INT       NOT NULL,
    started_at                      TIMESTAMP NOT NULL,
    ended_at                        TIMESTAMP NULL,
    created_at                      DATETIME  NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at                      DATETIME  NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    deleted_at                      DATETIME  NULL,
    PRIMARY KEY (maintenance_order_time_entry_id),
    CONSTRAINT fk_mote_order FOREIGN KEY (maintenance_order_id) REFERENCES maintenance_order                  (maintenance_order_id),
    CONSTRAINT fk_mote_type  FOREIGN KEY (type_id)              REFERENCES maintenance_order_time_entry_type  (maintenance_order_time_entry_type_id)
);

-- ── Schedule dates ────────────────────────────────────────────────────────────

CREATE TABLE IF NOT EXISTS maintenance_order_schedule_date (
    maintenance_order_schedule_date_id INT      NOT NULL AUTO_INCREMENT,
    maintenance_order_id               INT      NOT NULL,
    date                               DATETIME NOT NULL,
    created_at                         DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at                         DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    deleted_at                         DATETIME NULL,
    PRIMARY KEY (maintenance_order_schedule_date_id),
    CONSTRAINT fk_mosd_order FOREIGN KEY (maintenance_order_id) REFERENCES maintenance_order (maintenance_order_id)
);

CREATE TABLE IF NOT EXISTS maintenance_order_schedule_date_technician (
    maintenance_order_schedule_date_technician_id INT          NOT NULL AUTO_INCREMENT,
    maintenance_order_schedule_date_id            INT          NOT NULL,
    technician_id                                 INT          NOT NULL,
    technician_name                               VARCHAR(255) NOT NULL,
    created_at                                    DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at                                    DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    deleted_at                                    DATETIME     NULL,
    PRIMARY KEY (maintenance_order_schedule_date_technician_id),
    CONSTRAINT fk_mosdt_sd FOREIGN KEY (maintenance_order_schedule_date_id)
        REFERENCES maintenance_order_schedule_date (maintenance_order_schedule_date_id) ON DELETE CASCADE
);

-- ── Attachments ───────────────────────────────────────────────────────────────

CREATE TABLE IF NOT EXISTS maintenance_order_attachment (
    maintenance_order_attachment_id INT          NOT NULL AUTO_INCREMENT,
    maintenance_order_id            INT          NOT NULL,
    file_name                       VARCHAR(255) NOT NULL,
    file_path                       VARCHAR(500) NOT NULL,
    file_size                       INT          NOT NULL,
    created_at                      DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at                      DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    deleted_at                      DATETIME     NULL,
    PRIMARY KEY (maintenance_order_attachment_id),
    CONSTRAINT fk_moa_order FOREIGN KEY (maintenance_order_id) REFERENCES maintenance_order (maintenance_order_id)
);

-- ── Checklist templates ───────────────────────────────────────────────────────

CREATE TABLE IF NOT EXISTS maintenance_order_checklist_template (
    maintenance_order_checklist_template_id INT          NOT NULL AUTO_INCREMENT,
    name                                    VARCHAR(255) NOT NULL,
    reason_id                               INT          NULL,
    active                                  TINYINT(1)   NOT NULL DEFAULT 1,
    created_at                              DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at                              DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    deleted_at                              DATETIME     NULL,
    PRIMARY KEY (maintenance_order_checklist_template_id)
);

CREATE TABLE IF NOT EXISTS maintenance_order_checklist_item (
    maintenance_order_checklist_item_id INT          NOT NULL AUTO_INCREMENT,
    template_id                         INT          NOT NULL,
    section                             VARCHAR(255) NULL,
    label                               TEXT         NOT NULL,
    type                                ENUM('single_choice','multi_choice','text','table_row','sap_item') NOT NULL,
    options                             TEXT         NULL,
    metadata                            TEXT         NULL,
    required                            TINYINT(1)   NOT NULL DEFAULT 0,
    sort_order                          INT          NOT NULL DEFAULT 0,
    parent_item_id                      INT          NULL,
    show_when_values                    TEXT         NULL,
    created_at                          DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at                          DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    deleted_at                          DATETIME     NULL,
    PRIMARY KEY (maintenance_order_checklist_item_id),
    CONSTRAINT fk_moci_template FOREIGN KEY (template_id) REFERENCES maintenance_order_checklist_template (maintenance_order_checklist_template_id)
);

-- ── Order checklists ──────────────────────────────────────────────────────────

CREATE TABLE IF NOT EXISTS maintenance_order_checklist (
    maintenance_order_checklist_id INT      NOT NULL AUTO_INCREMENT,
    maintenance_order_id           INT      NOT NULL,
    template_id                    INT      NOT NULL,
    completed_at                   DATETIME NULL,
    created_at                     DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at                     DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    deleted_at                     DATETIME NULL,
    PRIMARY KEY (maintenance_order_checklist_id),
    CONSTRAINT fk_moc_order    FOREIGN KEY (maintenance_order_id) REFERENCES maintenance_order                    (maintenance_order_id),
    CONSTRAINT fk_moc_template FOREIGN KEY (template_id)          REFERENCES maintenance_order_checklist_template (maintenance_order_checklist_template_id)
);

CREATE TABLE IF NOT EXISTS maintenance_order_checklist_answer (
    maintenance_order_checklist_answer_id INT  NOT NULL AUTO_INCREMENT,
    checklist_id                          INT  NOT NULL,
    item_id                               INT  NOT NULL,
    value                                 TEXT NULL,
    created_at                            DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at                            DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    deleted_at                            DATETIME NULL,
    PRIMARY KEY (maintenance_order_checklist_answer_id),
    CONSTRAINT fk_moca_checklist FOREIGN KEY (checklist_id) REFERENCES maintenance_order_checklist      (maintenance_order_checklist_id),
    CONSTRAINT fk_moca_item      FOREIGN KEY (item_id)      REFERENCES maintenance_order_checklist_item  (maintenance_order_checklist_item_id)
);

-- ── Config ────────────────────────────────────────────────────────────────────

CREATE TABLE IF NOT EXISTS maintenance_order_config (
    maintenance_order_config_id INT NOT NULL AUTO_INCREMENT,
    PRIMARY KEY (maintenance_order_config_id)
);

CREATE TABLE IF NOT EXISTS maintenance_order_config_assignable_access_level (
    maintenance_order_config_id INT NOT NULL,
    access_level_id             INT NOT NULL,
    PRIMARY KEY (maintenance_order_config_id, access_level_id),
    FOREIGN KEY (maintenance_order_config_id) REFERENCES maintenance_order_config (maintenance_order_config_id) ON DELETE CASCADE,
    FOREIGN KEY (access_level_id)             REFERENCES access_level             (access_level_id)             ON DELETE CASCADE
);

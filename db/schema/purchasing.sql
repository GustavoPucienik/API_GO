-- ── Purchasing module ─────────────────────────────────────────────────────────

CREATE TABLE IF NOT EXISTS purchasing_order_status (
    purchasing_order_status_id INT          NOT NULL AUTO_INCREMENT,
    name                       VARCHAR(100) NOT NULL UNIQUE,
    background_color           VARCHAR(20)  NOT NULL,
    border_color               VARCHAR(20)  NOT NULL,
    text_color                 VARCHAR(20)  NOT NULL,
    sort_index                 INT          NOT NULL DEFAULT 0,
    is_closed                  BOOLEAN      NOT NULL DEFAULT FALSE,
    is_default                 BOOLEAN      NOT NULL DEFAULT FALSE,
    is_pending_approval        BOOLEAN      NOT NULL DEFAULT FALSE,
    PRIMARY KEY (purchasing_order_status_id)
);

CREATE TABLE IF NOT EXISTS purchasing_order_priorities (
    purchasing_order_priority_id INT          NOT NULL AUTO_INCREMENT,
    name                         VARCHAR(100) NOT NULL UNIQUE,
    background_color             VARCHAR(20)  NOT NULL,
    border_color                 VARCHAR(20)  NOT NULL,
    text_color                   VARCHAR(20)  NOT NULL,
    sort_index                   INT          NOT NULL DEFAULT 0,
    PRIMARY KEY (purchasing_order_priority_id)
);

CREATE TABLE IF NOT EXISTS purchasing_order_reason (
    purchasing_order_reason_id INT          NOT NULL AUTO_INCREMENT,
    name                       VARCHAR(100) NOT NULL UNIQUE,
    initial_status_id          INT          NOT NULL,
    created_at                 DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at                 DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    deleted_at                 DATETIME     NULL,
    PRIMARY KEY (purchasing_order_reason_id),
    FOREIGN KEY (initial_status_id) REFERENCES purchasing_order_status(purchasing_order_status_id)
);

CREATE TABLE IF NOT EXISTS purchasing_order_status_transition (
    id               INT NOT NULL AUTO_INCREMENT,
    from_status_id   INT NULL,
    to_status_id     INT NOT NULL,
    access_level_id  INT NOT NULL,
    PRIMARY KEY (id),
    UNIQUE KEY uq_purchasing_transition (from_status_id, to_status_id, access_level_id),
    FOREIGN KEY (from_status_id)  REFERENCES purchasing_order_status(purchasing_order_status_id),
    FOREIGN KEY (to_status_id)    REFERENCES purchasing_order_status(purchasing_order_status_id),
    FOREIGN KEY (access_level_id) REFERENCES access_level(access_level_id)
);

CREATE TABLE IF NOT EXISTS purchasing_order (
    purchasing_order_id  INT          NOT NULL AUTO_INCREMENT,
    requester_id         INT          NOT NULL,
    requester_name       VARCHAR(255) NOT NULL,
    assigned_to_id       INT          NULL,
    assigned_to_name     VARCHAR(255) NULL,
    salesperson_name     VARCHAR(255) NULL,
    client_code          VARCHAR(50)  NOT NULL,
    client_name          VARCHAR(255) NOT NULL,
    client_alias_name    VARCHAR(255) NOT NULL,
    status_id            INT          NOT NULL,
    priority_id          INT          NOT NULL,
    reason_id            INT          NOT NULL,
    description          TEXT         NOT NULL,
    captured_at          DATETIME     NULL,
    resolution           TEXT         NULL,
    closed_at            DATETIME     NULL,
    created_at           DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at           DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    deleted_at           DATETIME     NULL,
    PRIMARY KEY (purchasing_order_id),
    FOREIGN KEY (status_id)   REFERENCES purchasing_order_status(purchasing_order_status_id),
    FOREIGN KEY (priority_id) REFERENCES purchasing_order_priorities(purchasing_order_priority_id),
    FOREIGN KEY (reason_id)   REFERENCES purchasing_order_reason(purchasing_order_reason_id)
);

CREATE TABLE IF NOT EXISTS purchasing_order_attachment (
    purchasing_order_attachment_id INT          NOT NULL AUTO_INCREMENT,
    purchasing_order_id            INT          NOT NULL,
    file_name                      VARCHAR(255) NOT NULL,
    file_path                      VARCHAR(500) NOT NULL,
    file_size                      INT          NOT NULL,
    created_at                     DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at                     DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    deleted_at                     DATETIME     NULL,
    PRIMARY KEY (purchasing_order_attachment_id),
    FOREIGN KEY (purchasing_order_id) REFERENCES purchasing_order(purchasing_order_id) ON DELETE CASCADE
);

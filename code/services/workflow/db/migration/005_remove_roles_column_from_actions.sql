-- Migration 005: Remove roles column from actions table
-- Roles are now handled through attributeValidation.attributes

ALTER TABLE actions DROP COLUMN IF EXISTS roles; 
DELETE FROM users WHERE email = 'admin@flowforge.com';
DELETE FROM tenants WHERE slug = 'flowforge';

INSERT INTO tenants (id, name, slug) VALUES ('d8c47b56-7871-460f-968f-9a7e37e96b66', 'FlowForge Super Admin', 'superadmin') ON CONFLICT DO NOTHING;
INSERT INTO users (id, tenant_id, email, password_hash, role, is_active) VALUES ('f1a2b3c4-d5e6-4789-8a9b-0c1d2e3f4a5b', 'd8c47b56-7871-460f-968f-9a7e37e96b66', 'superadmin@flowforge.com', '$2a$10$Lcsp4KbT8ardCNPdBzYMWeH446IZVPSL9I88pK8vSZs55oG4R.hhq', 'super-admin', true) ON CONFLICT DO NOTHING;

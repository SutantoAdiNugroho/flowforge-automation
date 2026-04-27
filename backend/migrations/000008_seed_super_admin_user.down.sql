-- Delete super-admin user and tenant
DELETE FROM users WHERE email = 'superadmin@flowforge.com';
DELETE FROM tenants WHERE slug = 'superadmin';

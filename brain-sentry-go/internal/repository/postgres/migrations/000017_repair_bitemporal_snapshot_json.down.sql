-- Irreversible data repair. Restoring the incompatible database-shaped JSON
-- would reintroduce silent corruption in historical reads.
SELECT 1;

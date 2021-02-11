BEGIN;

CREATE OR REPLACE FUNCTION delete_async_unregistered_apps()
    RETURNS TRIGGER
AS
$$
BEGIN
    IF (NEW.error IS NULL AND NEW.ready=true AND NEW.deleted_at > NEW.created_at) THEN
        DELETE FROM applications
        WHERE id=NEW.id;
    END IF;

    RETURN NEW;
END
$$ LANGUAGE plpgsql;

CREATE TRIGGER cleanup_deleted_application_resource
    AFTER UPDATE ON applications
    FOR ROW
EXECUTE PROCEDURE delete_async_unregistered_apps();

COMMIT;
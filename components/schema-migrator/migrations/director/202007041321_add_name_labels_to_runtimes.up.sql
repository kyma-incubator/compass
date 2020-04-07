IF NOT EXISTS ( SELECT * FROM labels 
                   WHERE "key"='name'
                   AND runtime_id=id
                   )
BEGIN
    INSERT INTO labels
        (id ,tenant_id, runtime_id, "key","value")
    SELECT uuid_in(md5(random()::text || clock_timestamp()::text)::cstring), tenant_id, id, 'name', to_json("name")
    FROM runtimes 
END
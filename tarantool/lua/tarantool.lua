box.cfg {
    listen = 3301
}
box.once("bootstrap", function()
    s = box.schema.space.create('mysqldaemon')
    s:create_index('primary', {type = 'tree', parts = {1, 'unsigned'}, if_not_exists = true})

    t = box.schema.space.create('mysqldata')
    t:create_index('primary', {type = 'tree', parts = {1, 'unsigned'}, if_not_exists = true})
    t:create_index('user_idx', {type = 'tree', parts = {{2, 'string'}, {1, 'unsigned'}}, if_not_exists = true, unique = false})
end)
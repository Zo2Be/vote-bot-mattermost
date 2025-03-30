box.cfg { listen = 3301 }

box.schema.user.create('bot', {
    password = '123',
    if_not_exists = true
})

if box.space.polls == nil then
    box.schema.space.create('polls', {
        if_not_exists = true,
        format = {
            { name = 'id', type = 'string' },
            { name = 'creator', type = 'string' },
            { name = 'question', type = 'string' },
            { name = 'options', type = 'array' },
            { name = 'votes', type = 'map' },
            { name = 'active', type = 'boolean' },
            { name = 'created_at', type = 'unsigned' },
            { name = 'closed_at', type = 'unsigned', is_nullable = true }
        }
    })
    box.space.polls:create_index('primary', {
        parts = { 'id' },
        if_not_exists = true
    })
end

box.schema.user.grant('bot', 'read,write', 'space', 'polls')
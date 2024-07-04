package workqueue

import "github.com/redis/go-redis/v9"

var allowN = redis.NewScript(`

-- KEYS[1] as tokens_key
-- KEYS[2] as timestamp_key

-- produce token produce rate
local rate = tonumber(ARGV[1])
-- bucket capacity
local capacity = tonumber(ARGV[2])
-- current time
local now = tonumber(ARGV[3])
-- consumed token number
local consumed = tonumber(ARGV[4])

-- how long the bucket will be full, in milliseconds
local fill_time = 1000*capacity/rate
-- how long the token will be expired, 2 times of fill_time
local ttl = math.floor(fill_time*2)
-- current token number
local current_tokens = tonumber(redis.call("get", KEYS[1]))
-- if current_tokens is nil, set it to capacity
if current_tokens == nil then
    current_tokens = capacity
end

-- last refresh time
local last_refreshed = tonumber(redis.call("get", KEYS[2]))
-- if last_refreshed is nil, set it to 0
if last_refreshed == nil then
    last_refreshed = 0
end

-- calculate the new token number
local delta = math.max(0, now-last_refreshed)
local filled_tokens = math.min(capacity, current_tokens+(delta*rate/1000))
-- check if the token is enough
local allowed = filled_tokens >= consumed
local new_tokens = filled_tokens
if allowed then
    new_tokens = filled_tokens - consumed
end

redis.call("PSETEX", KEYS[1], ttl, new_tokens)
redis.call("PSETEX", KEYS[2], ttl, now)

return allowed
`)

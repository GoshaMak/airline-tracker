local email = redis.call("GET", KEYS[1])
if not email then
	return 0
end
if redis.call("EXISTS", KEYS[3]) == 1 then
	return -3
end
local stored = redis.call("GET", KEYS[2])
if not stored then
	return 0
end
if stored ~= ARGV[1] then
	local attempts = redis.call("INCR", KEYS[4])
	if attempts == 1 then
		redis.call("EXPIRE", KEYS[4], ARGV[3])
	end
	if attempts >= tonumber(ARGV[2]) then
		redis.call("SET", KEYS[3], "1", "EX", ARGV[4])
		return -3
	end
	return -2
end
redis.call("DEL", KEYS[1], KEYS[2], KEYS[4])
return 1

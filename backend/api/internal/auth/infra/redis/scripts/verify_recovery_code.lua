if redis.call("EXISTS", KEYS[2]) == 0 then
	return 0
end
local stored = redis.call("GET", KEYS[1])
if not stored then
	return 0
end
if stored ~= ARGV[1] then
	local attempts = redis.call("INCR", KEYS[4])
	if attempts == 1 then
		redis.call("EXPIRE", KEYS[4], ARGV[3])
	end
	if attempts >= tonumber(ARGV[2]) then
		redis.call("DEL", KEYS[1])
	end
	return -1
end
redis.call("DEL", KEYS[1], KEYS[2], KEYS[3], KEYS[4])
return 1

local attempts = redis.call("INCR", KEYS[1])
if attempts == 1 then
	redis.call("EXPIRE", KEYS[1], ARGV[2])
end
if attempts >= tonumber(ARGV[1]) then
	redis.call("SET", KEYS[2], "1", "EX", ARGV[3])
	return 1
end
return 0

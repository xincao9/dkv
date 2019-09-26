wrk.method="POST"
wrk.headers["content-type"]="application/json"

request = function()
	local kv = string.format('{"k":"%d", "v":"%d"}', math.random(1, 100000000), math.random(1, 100000000))
	return wrk.format(nil, nil, nil, kv)
end

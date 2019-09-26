wrk.method="GET"

request = function()
	local path = string.format("/kv/%d", math.random(1, 100000000))
	return wrk.format(nil, path, nil, nil)
end

response = function(status, headers, body)
	print(status, body)
end

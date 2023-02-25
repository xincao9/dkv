wrk.method="DELETE"

math.randomseed(os.time())

request = function()
	local path = string.format("/kv/%d", math.random(2000000))
	return wrk.format(nil, path, nil, nil)
end

response = function(status, headers, body)
	print(status, body)
end

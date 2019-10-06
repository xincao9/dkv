wrk.method="GET"

count = 1

request = function()
	local path = string.format("/kv/%d", count)
	count = count + 1
	return wrk.format(nil, path, nil, nil)
end

--response = function(status, headers, body)
--	print(status, body)
--end

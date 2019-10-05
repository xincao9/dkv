wrk.method="PUT"
wrk.headers["content-type"]="application/json"

request = function()
	local kv = string.format('{"k":"%d", "v":"%s"}', math.random(1, 100000000), "The filepath package provides functions to parse and construct file paths in a way that is portable between operating systems; dir/file on Linux vs. dir file on Windows, for example")
	return wrk.format(nil, nil, nil, kv)
end

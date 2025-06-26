package locator

const (

	// 解绑网关脚本
	unbindGateScript = `
	local val = redis.call('GET', KEYS[1])

	if not val or val ~= ARGV[1] then
		return {'NO'}
	end

	redis.call('DEL', KEYS[1])

	return {'OK'}
`

	// 解绑游戏脚本
	unbindGameScript = `
	local val = redis.call('HGET', KEYS[1], ARGV[1])

	if not val or val ~= ARGV[2] then
		return {'NO'}
	end

	redis.call('HDEL', KEYS[1], ARGV[1])

	return {'OK'}
`
)

import { env } from '$env/dynamic/private';

export const backendApiPrefix = () => {
	return env.API_PREFIX || 'http://0.0.0.0:3000'
}


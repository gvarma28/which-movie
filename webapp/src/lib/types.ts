export interface Movie {
	movie_name: string;
	year: number;
	short_description: string;
}

export interface IsValidUrlReponse {
	isValid: boolean;
	url: string;
}

export interface MagicResponse {
	success: boolean;
	results: Movie[]
}
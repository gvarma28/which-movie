export interface Movie {
	title: string;
	year: number;
	description: string;
}

export interface IsValidUrlReponse {
	isValid: boolean;
	url: string;
}
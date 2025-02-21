/// <reference types="vite/client" />

interface ImportMetaEnv {
	readonly VITE_SERVER_URL: string;
	readonly VITE_TESTING_UI: string;
	// Add other environment variables here as needed
  }
  
  interface ImportMeta {
	readonly env: ImportMetaEnv;
  }
  
export type Wall = {
	id: number;
	user_id: number;
	title: string;
	description: string;
	is_public: boolean;
	is_archived: boolean;
	is_deleted: boolean;
	popularity_score: number;
	created_at: string;
	updated_at: string;
};

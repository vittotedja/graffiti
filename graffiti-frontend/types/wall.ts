export type Wall = {
	id: string;
	user_id: string;
	title: string;
	description: string;
	background_image: string;
	is_public: boolean;
	is_archived: boolean;
	is_deleted: boolean;
	is_pinned: boolean;
	popularity_score: number;
	created_at: string;
	updated_at: string;
};

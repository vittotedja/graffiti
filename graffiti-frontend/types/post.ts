export type Post = {
	id: string;
	wall_id: string;
	author: string;
	media_url: string;
	post_type: "media" | "embed_url";
	is_highlighted: boolean;
	likes_count: number;
	is_deleted: boolean;
	created_at: string;
	profile_picture: string;
	username: string;
	fullname: string;
};

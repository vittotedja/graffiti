export type Post = {
	id: string;
	wall_id: string;
	author: string;
	media_url: string;
	post_type: "media" | "embed_link";
	caption: string;
	is_highlighted: boolean;
	likes_count: number;
	is_deleted: boolean;
	created_at: string;
	profile_picture: string;
	username: string;
	fullname: string;
};

export type RequestPost = {
	wall_id: string;
	media_url: string | null;
	post_type: "media" | "embed_link";
	// caption: string;
};

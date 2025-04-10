export type User = {
	id: string;
	username: string;
	fullname: string;
	email: string;
	profile_picture: string;
	bio: string;
	has_onboarded: boolean;
	background_image: string;
	onboarding_at: Date;
	createdAt: Date;
	updatedAt: Date;
};

export type UserWithMutualFriends = User & {
	mutual_friend_count: number;
};

export type Friendship = {
	ID: string;
	UserID: string;
	ToUser: string;
	FromUser: string;
	Status: {
		Status: "pending" | "friends" | "blocked";
		Valid: boolean;
	};
	Fullname: string;
	Username: string;
	ProfilePicture: string;
	CreatedAt: string;
	UpdatedAt: string;
};

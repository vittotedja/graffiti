import {useState, useRef} from "react";
import {Avatar, AvatarFallback, AvatarImage} from "@/components/ui/avatar";
import {
	Dialog,
	DialogContent,
	DialogHeader,
	DialogTitle,
	DialogFooter,
} from "@/components/ui/dialog";
import {Button} from "@/components/ui/button";
import {Input} from "@/components/ui/input";
import {Textarea} from "@/components/ui/textarea";
import {Camera, ImageIcon, Loader2, X} from "lucide-react";
import {formatFullName} from "@/lib/formatter";
import {toast} from "sonner";
import {cn} from "@/lib/utils";
import {getPresignedUrl, uploadToS3} from "@/lib/s3-uploader";
import {fetchWithAuth} from "@/lib/auth";
import Image from "next/image";

interface EditProfileModalProps {
	isOpen: boolean;
	onClose: () => void;
	user: {
		fullname: string;
		username: string;
		bio?: string;
		profile_picture?: string;
		background_image?: string;
		id: string;
	};
	onSave: (userData: {
		fullname: string;
		username: string;
		bio: string;
		avatar?: File;
		backgroundImage?: File;
	}) => void;
}

export function EditProfileModal({
	isOpen,
	onClose,
	user,
}: EditProfileModalProps) {
	const [fullname, setFullname] = useState(user.fullname);
	const [username, setUsername] = useState(user.username);
	const [bio, setBio] = useState(user.bio || "");
	const [avatarFile, setAvatarFile] = useState<File | null>(null);
	const [backgroundFile, setBackgroundFile] = useState<File | null>(null);
	const [avatarPreview, setAvatarPreview] = useState<string | null>(null);
	const [backgroundPreview, setBackgroundPreview] = useState<string | null>(
		null
	);
	const [loading, setLoading] = useState(false);

	const avatarInputRef = useRef<HTMLInputElement>(null);
	const backgroundInputRef = useRef<HTMLInputElement>(null);

	const handleAvatarChange = (e: React.ChangeEvent<HTMLInputElement>) => {
		if (e.target.files && e.target.files[0]) {
			const file = e.target.files[0];
			setAvatarFile(file);
			setAvatarPreview(URL.createObjectURL(file));
		}
	};

	const handleBackgroundChange = (e: React.ChangeEvent<HTMLInputElement>) => {
		if (e.target.files && e.target.files[0]) {
			const file = e.target.files[0];
			setBackgroundFile(file);
			setBackgroundPreview(URL.createObjectURL(file));
		}
	};

	const handleSubmit = async (e: React.FormEvent) => {
		e.preventDefault();
		setLoading(true);

		try {
			let avatarUrl = user.profile_picture;
			let backgroundUrl = user.background_image;

			if (avatarFile) {
				const presignedUrlAvatarData = await getPresignedUrl(
					`${user.id}.png`,
					avatarFile,
					"profile"
				);
				await uploadToS3(presignedUrlAvatarData.presignedUrl, avatarFile);
				avatarUrl = presignedUrlAvatarData.publicUrl;
			}
			if (backgroundFile) {
				const presignedUrlBackgroundData = await getPresignedUrl(
					`${user.id}.png`,
					backgroundFile,
					"background_image"
				);
				await uploadToS3(
					presignedUrlBackgroundData.presignedUrl,
					backgroundFile
				);
				backgroundUrl = presignedUrlBackgroundData.publicUrl;
			}

			// Update DB
			const response = await fetchWithAuth("/api/v2/users", {
				method: "POST",
				body: JSON.stringify({
					username,
					fullname,
					profile_picture: avatarUrl,
					background_image: backgroundUrl,
					bio,
				}),
			});
			if (!response.ok) throw new Error("Profile failed to update");

			toast("Profile updated", {
				description:
					"Your profile has been updated successfully. Changes may take a few minutes to reflect.",
			});

			onClose();
		} catch (e) {
			console.error(e);
			toast.error("Failed to update your profile. Please try again.");
		} finally {
			setLoading(false);
		}
	};

	const resetForm = () => {
		setFullname(user.fullname);
		setUsername(user.username);
		setBio(user.bio || "");
		setAvatarFile(null);
		setBackgroundFile(null);
		setAvatarPreview(null);
		setBackgroundPreview(null);
	};

	return (
		<Dialog
			open={isOpen}
			onOpenChange={(open) => {
				if (!open) {
					resetForm();
					onClose();
				}
			}}
		>
			<DialogContent className="sm:max-w-[600px] max-h-[90vh] overflow-y-auto">
				<DialogHeader>
					<DialogTitle className="text-xl font-bold">
						Edit Your Profile
					</DialogTitle>
				</DialogHeader>

				<form onSubmit={handleSubmit} className="space-y-6 py-4">
					{/* Background Image Section */}
					<div className="relative w-full h-48 rounded-lg overflow-hidden bg-muted">
						{backgroundPreview ? (
							<Image
								src={backgroundPreview}
								alt="Background preview"
								className="w-full h-full object-cover"
								width={"1000"}
								height={"400"}
							/>
						) : user.background_image ? (
							<Image
								src={user.background_image}
								alt="Background"
								className="w-full h-full object-cover"
								width={"1000"}
								height={"400"}
							/>
						) : (
							<div className="absolute inset-0 flex items-center justify-center bg-muted">
								<ImageIcon className="h-12 w-12 text-muted-foreground opacity-50" />
							</div>
						)}

						<Button
							type="button"
							variant="secondary"
							size="sm"
							className="absolute bottom-3 right-3"
							onClick={() => backgroundInputRef.current?.click()}
						>
							<Camera className="h-4 w-4 mr-2" />
							Change Cover
						</Button>

						{(backgroundPreview || user.background_image) && (
							<Button
								type="button"
								variant="destructive"
								size="icon"
								className="absolute top-3 right-3 h-8 w-8"
								onClick={() => {
									setBackgroundFile(null);
									setBackgroundPreview(null);
								}}
							>
								<X className="h-4 w-4" />
							</Button>
						)}

						<input
							ref={backgroundInputRef}
							type="file"
							accept="image/*"
							className="hidden"
							onChange={handleBackgroundChange}
						/>
					</div>

					{/* Avatar Section */}
					<div className="flex items-center space-x-4">
						<div className="relative">
							<Avatar className="h-24 w-24 border-4 border-background">
								<AvatarImage
									src={
										avatarPreview ||
										user.profile_picture ||
										"/placeholder.svg?height=96&width=96"
									}
									alt="User Avatar"
								/>
								<AvatarFallback>{formatFullName(fullname)}</AvatarFallback>
							</Avatar>

							<Button
								type="button"
								variant="secondary"
								size="icon"
								className="absolute bottom-0 right-0 h-8 w-8 rounded-full"
								onClick={() => avatarInputRef.current?.click()}
							>
								<Camera className="h-4 w-4" />
							</Button>

							<input
								ref={avatarInputRef}
								type="file"
								accept="image/*"
								className="hidden"
								onChange={handleAvatarChange}
							/>
						</div>

						<div className="flex-1">
							<h3 className="text-lg font-medium">Profile Picture</h3>
							<p className="text-sm text-muted-foreground">
								Upload a new avatar image. Square images work best.
							</p>
						</div>
					</div>

					{/* Personal Details */}
					<div className="space-y-4">
						<div>
							<label
								htmlFor="fullname"
								className="block text-sm font-medium mb-1"
							>
								Full Name
							</label>
							<Input
								id="fullname"
								value={fullname}
								onChange={(e) => setFullname(e.target.value)}
								required
							/>
						</div>

						<div>
							<label
								htmlFor="username"
								className="block text-sm font-medium mb-1"
							>
								Username
							</label>
							<div className="flex">
								<div className="bg-muted flex items-center px-3 rounded-l-md border border-r-0 border-input">
									<span className="text-muted-foreground">@</span>
								</div>
								<Input
									id="username"
									value={username}
									onChange={(e) => setUsername(e.target.value)}
									className="rounded-l-none"
									required
								/>
							</div>
						</div>

						<div>
							<label htmlFor="bio" className="block text-sm font-medium mb-1">
								Bio
							</label>
							<Textarea
								id="bio"
								value={bio}
								onChange={(e) => setBio(e.target.value)}
								placeholder="Tell us about yourself..."
								className="resize-none min-h-[100px]"
							/>
						</div>
					</div>

					<DialogFooter
						className={cn("sm:justify-end", {
							"opacity-50 pointer-events-none": loading,
						})}
					>
						<Button type="button" variant="outline" onClick={onClose}>
							Cancel
						</Button>
						<Button type="submit" disabled={loading}>
							{loading && <Loader2 className="mr-2 h-4 w-4 animate-spin" />}
							Save Changes
						</Button>
					</DialogFooter>
				</form>
			</DialogContent>
		</Dialog>
	);
}

import {useState, useRef, useCallback} from "react";
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
import {Camera, Check, CropIcon, ImageIcon, Loader2, X} from "lucide-react";
import {formatFullName} from "@/lib/formatter";
import {toast} from "sonner";
import {cn} from "@/lib/utils";
import {getPresignedUrl, uploadToS3} from "@/lib/s3-uploader";
import {fetchWithAuth} from "@/lib/auth";
import Image from "next/image";
import ReactCrop, {type Crop, type PixelCrop} from "react-image-crop";
import "react-image-crop/dist/ReactCrop.css";

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
}

// Constants for aspect ratios
const AVATAR_ASPECT_RATIO = 1; // 1:1 square
const BACKGROUND_ASPECT_RATIO = 3; // 3:1 banner (typical cover photo ratio)

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
	const [isAvatarRemoved, setIsAvatarRemoved] = useState(false);
	const [isBackgroundRemoved, setIsBackgroundRemoved] = useState(false);
	const [loading, setLoading] = useState(false);

	const [avatarCrop, setAvatarCrop] = useState<Crop>({
		unit: "%",
		width: 100,
		height: 100,
		x: 0,
		y: 0,
		// aspect: AVATAR_ASPECT_RATIO,
	});
	const [backgroundCrop, setBackgroundCrop] = useState<Crop>({
		unit: "%",
		width: 100,
		height: 33.33,
		x: 0,
		y: 0,
		// aspect: BACKGROUND_ASPECT_RATIO,
	});

	const [isCroppingAvatar, setIsCroppingAvatar] = useState(false);
	const [isCroppingBackground, setIsCroppingBackground] = useState(false);

	const avatarImgRef = useRef<HTMLImageElement>(null);
	const backgroundImgRef = useRef<HTMLImageElement>(null);
	const avatarInputRef = useRef<HTMLInputElement>(null);
	const backgroundInputRef = useRef<HTMLInputElement>(null);

	const handleAvatarChange = (e: React.ChangeEvent<HTMLInputElement>) => {
		if (e.target.files && e.target.files[0]) {
			const file = e.target.files[0];
			const previewUrl = URL.createObjectURL(file);
			setAvatarPreview(previewUrl);
			setIsAvatarRemoved(false);
			setIsCroppingAvatar(true);
		}
	};

	const handleBackgroundChange = (e: React.ChangeEvent<HTMLInputElement>) => {
		if (e.target.files && e.target.files[0]) {
			const file = e.target.files[0];
			const previewUrl = URL.createObjectURL(file);
			setBackgroundPreview(previewUrl);
			setIsBackgroundRemoved(false);
			setIsCroppingBackground(true);
		}
	};

	const cropImage = useCallback(
		async (
			image: HTMLImageElement | null,
			crop: PixelCrop,
			fileName: string
		): Promise<File | null> => {
			if (!image) return null;

			// Get the device pixel ratio
			// const pixelRatio = window.devicePixelRatio || 1;

			// Calculate the canvas size based on the original image dimensions
			// This preserves the full resolution of the cropped region
			const scaleX = image.naturalWidth / image.width;
			const scaleY = image.naturalHeight / image.height;

			const canvas = document.createElement("canvas");
			// Use the full resolution dimensions of the cropped area
			canvas.width = crop.width * scaleX;
			canvas.height = crop.height * scaleY;

			const ctx = canvas.getContext("2d");
			if (!ctx) return null;

			// Set canvas properties for better quality
			ctx.imageSmoothingEnabled = true;
			ctx.imageSmoothingQuality = "high";

			// Draw the cropped portion of the image at full resolution
			ctx.drawImage(
				image,
				crop.x * scaleX,
				crop.y * scaleY,
				crop.width * scaleX,
				crop.height * scaleY,
				0,
				0,
				canvas.width,
				canvas.height
			);

			// Create a high-quality image file from the canvas
			return new Promise((resolve) => {
				canvas.toBlob(
					(blob) => {
						if (!blob) {
							resolve(null);
							return;
						}
						const file = new File([blob], fileName, {type: "image/jpeg"});
						resolve(file);
					},
					"image/jpeg", // JPEG often works better for photos
					0.95 // High quality but with minimal compression
				);
			});
		},
		[]
	);

	const completeAvatarCrop = useCallback(async () => {
		if (!avatarImgRef.current) return;

		try {
			const croppedFile = await cropImage(
				avatarImgRef.current,
				avatarCrop as PixelCrop,
				`avatar-${Date.now()}.png`
			);

			if (croppedFile) {
				setAvatarFile(croppedFile);
				setAvatarPreview(URL.createObjectURL(croppedFile));
			}

			setIsCroppingAvatar(false);
		} catch (error) {
			console.error("Error cropping avatar:", error);
			toast.error("Failed to crop image");
		}
	}, [avatarCrop, cropImage]);

	const completeBackgroundCrop = useCallback(async () => {
		if (!backgroundImgRef.current) return;

		try {
			const croppedFile = await cropImage(
				backgroundImgRef.current,
				backgroundCrop as PixelCrop,
				`background-${Date.now()}.png`
			);

			if (croppedFile) {
				setBackgroundFile(croppedFile);
				setBackgroundPreview(URL.createObjectURL(croppedFile));
			}

			setIsCroppingBackground(false);
		} catch (error) {
			console.error("Error cropping background:", error);
			toast.error("Failed to crop image");
		}
	}, [backgroundCrop, cropImage]);

	const handleSubmit = async (e: React.FormEvent) => {
		e.preventDefault();
		setLoading(true);

		try {
			let avatarUrl = user.profile_picture;
			let backgroundUrl = user.background_image;

			if (isAvatarRemoved) {
				avatarUrl = "";
			} else if (avatarFile) {
				const presignedUrlAvatarData = await getPresignedUrl(
					`${user.id}.png`,
					avatarFile,
					"profile"
				);
				await uploadToS3(presignedUrlAvatarData.presignedUrl, avatarFile);
				avatarUrl = presignedUrlAvatarData.publicUrl;
			}

			// Handle background changes
			if (isBackgroundRemoved) {
				backgroundUrl = "";
			} else if (backgroundFile) {
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

	const showBackgroundImage = () => {
		if (isBackgroundRemoved) return false;
		if (backgroundPreview) return true;
		return !!user.background_image;
	};

	const getBackgroundImageSource = () => {
		if (backgroundPreview) return backgroundPreview;
		return user.background_image;
	};

	const getAvatarImageSource = () => {
		if (isAvatarRemoved) return "/placeholder.svg?height=96&width=96";
		if (avatarPreview) return avatarPreview;
		return user.profile_picture || "/placeholder.svg?height=96&width=96";
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
						{showBackgroundImage() ? (
							<Image
								src={getBackgroundImageSource() || ""}
								alt="Background"
								className="w-full h-full object-cover"
								width={1000}
								height={400}
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

						{backgroundPreview && !isBackgroundRemoved && (
							<Button
								type="button"
								variant="secondary"
								size="sm"
								className="absolute bottom-3 right-[160px]"
								onClick={() => setIsCroppingBackground(true)}
							>
								<CropIcon className="h-4 w-4" />
							</Button>
						)}

						{(backgroundPreview || user.background_image) && (
							<Button
								type="button"
								variant="destructive"
								size="icon"
								className="absolute top-3 right-3 h-8 w-8"
								onClick={() => {
									setBackgroundFile(null);
									setBackgroundPreview(null);
									setIsBackgroundRemoved(true);
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

					{isCroppingBackground && (
						<div className="fixed inset-0 bg-black/50 z-50 flex flex-col items-center justify-center p-4">
							<div className="bg-background p-4 rounded-lg max-w-4xl w-full">
								<h3 className="text-lg font-medium mb-2">Crop Cover Image</h3>
								<ReactCrop
									crop={backgroundCrop}
									onChange={(c) => setBackgroundCrop(c)}
									aspect={BACKGROUND_ASPECT_RATIO}
									className="max-h-[400px] max-w-full mx-auto"
								>
									{/* eslint-disable-next-line @next/next/no-img-element */}
									<img
										ref={backgroundImgRef}
										src={backgroundPreview || ""}
										alt="Crop background"
										className="max-h-[400px] max-w-full"
									/>
								</ReactCrop>
								<div className="flex gap-2 mt-4 justify-end">
									<Button
										type="button"
										onClick={completeBackgroundCrop}
										className="bg-green-600 hover:bg-green-700"
									>
										<Check className="h-4 w-4 mr-1" /> Apply Crop
									</Button>
									<Button
										type="button"
										variant="destructive"
										onClick={() => setIsCroppingBackground(false)}
									>
										<X className="h-4 w-4 mr-1" /> Cancel
									</Button>
								</div>
							</div>
						</div>
					)}

					{/* Avatar Section */}
					<div className="flex items-center space-x-4">
						<div className="relative">
							<Avatar className="h-24 w-24 border-4 border-background">
								<AvatarImage src={getAvatarImageSource()} alt="User Avatar" />
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

							{avatarPreview && !isAvatarRemoved && (
								<Button
									type="button"
									variant="secondary"
									size="icon"
									className="absolute bottom-0 left-0 h-8 w-8 rounded-full"
									onClick={() => setIsCroppingAvatar(true)}
								>
									<CropIcon className="h-4 w-4" />
								</Button>
							)}

							{(avatarPreview ||
								(user.profile_picture && !isAvatarRemoved)) && (
								<Button
									type="button"
									variant="destructive"
									size="icon"
									className="absolute top-0 right-0 h-6 w-6 rounded-full"
									onClick={() => {
										setAvatarFile(null);
										setAvatarPreview(null);
										setIsAvatarRemoved(true);
									}}
								>
									<X className="h-3 w-3" />
								</Button>
							)}

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

					{isCroppingAvatar && (
						<div className="fixed inset-0 bg-black/50 z-50 flex flex-col items-center justify-center p-4">
							<div className="bg-background p-4 rounded-lg max-w-md w-full">
								<h3 className="text-lg font-medium mb-2">
									Crop Profile Picture
								</h3>
								<ReactCrop
									crop={avatarCrop}
									onChange={(c) => setAvatarCrop(c)}
									aspect={AVATAR_ASPECT_RATIO}
									circularCrop
									className="max-h-[300px] max-w-full mx-auto"
								>
									{/* eslint-disable-next-line @next/next/no-img-element */}
									<img
										ref={avatarImgRef}
										src={avatarPreview || ""}
										alt="Crop avatar"
										className="max-h-[300px] max-w-full"
									/>
								</ReactCrop>
								<div className="flex gap-2 mt-4 justify-end">
									<Button
										type="button"
										onClick={completeAvatarCrop}
										className="bg-green-600 hover:bg-green-700"
									>
										<Check className="h-4 w-4 mr-1" /> Apply Crop
									</Button>
									<Button
										type="button"
										variant="destructive"
										onClick={() => setIsCroppingAvatar(false)}
									>
										<X className="h-4 w-4 mr-1" /> Cancel
									</Button>
								</div>
							</div>
						</div>
					)}

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
									readOnly
									className="rounded-l-none"
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

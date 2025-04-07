"use client";

import React from "react";
import {Button} from "@/components/ui/button";
import {Save} from "lucide-react";
import {Label} from "@/components/ui/label";
import {Textarea} from "@/components/ui/textarea";

interface DetailsTabProps {
	compositeDataUrl: string | null;
	caption: string;
	setCaption: (caption: string) => void;
	handleSubmit: () => void;
	isSubmitting: boolean;
}

const DetailsTab: React.FC<DetailsTabProps> = ({
	compositeDataUrl,
	caption,
	setCaption,
	handleSubmit,
	isSubmitting,
}) => {
	return (
		<div className="grid gap-6 md:grid-cols-[1fr_1fr]">
			<div className="relative border rounded-md overflow-hidden bg-white">
				{compositeDataUrl ? (
					// eslint-disable-next-line @next/next/no-img-element
					<img
						src={compositeDataUrl}
						alt="Combined Preview"
						className="max-w-full h-auto"
					/>
				) : (
					<p>Loading preview...</p>
				)}
			</div>
			<div className="space-y-4">
				<div className="space-y-2">
					<Label htmlFor="caption">Caption</Label>
					<Textarea
						id="caption"
						placeholder="Write a caption for your post..."
						value={caption}
						onChange={(e) => setCaption(e.target.value)}
						className="resize-none h-32"
					/>
				</div>
				<div className="pt-4">
					<Button
						className="w-full"
						onClick={handleSubmit}
						disabled={isSubmitting}
					>
						{isSubmitting ? (
							<div className="flex items-center gap-2">
								<div className="h-4 w-4 rounded-full border-2 border-t-transparent border-white animate-spin"></div>
								Posting...
							</div>
						) : (
							<>
								<Save className="h-4 w-4 mr-2" />
								Post to Wall
							</>
						)}
					</Button>
				</div>
			</div>
		</div>
	);
};

export default DetailsTab;

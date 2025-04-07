import {fetchWithAuth} from "./auth";

type PresignedUrlDataType = {
	key: string;
	presignedUrl: string;
	publicUrl: string;
};

/**
 * Converts an HTMLCanvasElement to a Blob.
 * @param canvas - The canvas element to convert.
 * @returns A promise that resolves with the Blob.
 */
export function canvasToBlob(canvas: HTMLCanvasElement): Promise<Blob> {
	return new Promise((resolve, reject) => {
		canvas.toBlob((blob) => {
			if (blob) {
				resolve(blob);
			} else {
				reject(new Error("Failed to generate image from canvas."));
			}
		}, "image/png");
	});
}

/**
 * Gets a presigned URL from the backend.
 * @param filename - The name for the file to upload.
 * @param file - The Blob representing the image.
 * @returns A promise that resolves with the presigned URL data.
 */
export async function getPresignedUrl(
	filename: string,
	file: Blob,
	upload_type: string = "uploads"
): Promise<PresignedUrlDataType> {
	console.log(upload_type);
	const requestBody = {
		filename,
		content_type: file.type,
		file_size: file.size,
		upload_type,
	};

	const response = await fetchWithAuth("/api/v1/presign", {
		method: "POST",
		headers: {
			"Content-Type": "application/json",
		},
		body: JSON.stringify(requestBody),
	});

	if (!response.ok) {
		throw new Error("Failed to get presigned URL");
	}

	const data = await response.json();
	// Assume the response contains the presigned URL as data.presignedUrl
	return data;
}

/**
 * Uploads a Blob to S3 using the provided presigned URL.
 * @param presignedUrl - The presigned S3 URL.
 * @param file - The Blob (e.g., image data).
 * @returns A promise that resolves when the upload is complete.
 */
export async function uploadToS3(
	presignedUrl: string,
	file: Blob
): Promise<Response> {
	const response = await fetch(presignedUrl, {
		method: "PUT",
		headers: {
			"Content-Type": file.type,
		},
		body: file,
	});
	if (!response.ok) {
		throw new Error("Failed to upload image to S3");
	}
	return response;
}

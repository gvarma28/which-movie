import axios from 'axios';

export interface Comment {
	comment: string;
	comment_id: string;
	continuationCommand: string | null;
}

export interface Comment {
	comment: string;
	comment_id: string;
	continuationCommand: string | null;
}

export interface GetCommentReponse {
	response: any;
	comment_info: Comment[];
	nextContinuationCommand: string | null;
}

export const initialRequest = async (url: string): Promise<any> => {
	try {
		let config = {
			method: 'GET',
			maxBodyLength: Infinity,
			url,
			headers: {
        'Origin': 'www.youtube.com',
        'Referer': 'www.youtube.com'
			}
		};
		const response = await axios.request(config);
		if (!response || response.status !== 200) {
			throw new Error(`Invalid Response Received -> status: ${response.status}`);
		}
		const responseStr = response?.data
			?.replaceAll('\n', '')
			?.split('var ytInitialData = ')?.[1]
			?.split(';</script>')?.[0];
		return JSON.parse(responseStr);
	} catch (error) {
		console.log('Error -> initialRequest');
		return null;
	}
};

export const getComments = async (
	continuationCommand: string
): Promise<GetCommentReponse | null> => {
	try {
		let data = JSON.stringify({
			context: {
				client: {
					clientName: 'WEB',
					clientVersion: '2.20240731.04.00'
				}
			},
			continuation: continuationCommand
		});
		const config = {
			method: 'post',
			maxBodyLength: Infinity,
			url: 'https://www.youtube.com/youtubei/v1/browse?prettyPrint=false',
			data: data,
			headers: {
				'Content-Type': 'application/json',
        'Origin': 'www.youtube.com',
        'Referer': 'www.youtube.com'
			}
		};
		const response = await axios.request(config);
		if (!response || response.status !== 200) {
			throw new Error(`Invalid Response Received -> status: ${response.status}`);
		}

		const continuations =
			response?.data?.onResponseReceivedEndpoints[1]?.reloadContinuationItemsCommand
				?.continuationItems ||
			response?.data?.onResponseReceivedEndpoints[0]?.appendContinuationItemsAction
				?.continuationItems ||
			null;

		continuationCommand =
			continuations?.[20]?.continuationItemRenderer?.continuationEndpoint?.continuationCommand
				?.token || null;
		// filter to have only comments elements
		const commentEntityPayload = response.data.frameworkUpdates.entityBatchUpdate.mutations.filter(
			(e: any) => e.payload.commentEntityPayload
		);
		// further map to have only comments elements
		const commentsObject = commentEntityPayload.map(
			(e: any) => e.payload.commentEntityPayload.properties
		);

		const comments: Comment[] = [];
		commentsObject.forEach((e: any) => {
			let contCommand;
			if (response?.data?.onResponseReceivedEndpoints?.[1]) {
				response?.data?.onResponseReceivedEndpoints?.[1]?.reloadContinuationItemsCommand?.continuationItems.forEach(
					(j: any) => {
						if (
							j?.commentThreadRenderer?.replies?.commentRepliesRenderer?.targetId?.includes(
								e.commentId
							)
						) {
							contCommand =
								j?.commentThreadRenderer?.replies?.commentRepliesRenderer?.contents?.[0]
									?.continuationItemRenderer?.continuationEndpoint?.continuationCommand?.token;
							return;
						}
					}
				);
			} else {
				response?.data?.onResponseReceivedEndpoints?.[0]?.appendContinuationItemsAction?.continuationItems.forEach(
					(j: any) => {
						if (
							j?.commentThreadRenderer?.replies?.commentRepliesRenderer?.targetId?.includes(
								e.commentId
							)
						) {
							contCommand =
								j?.commentThreadRenderer?.replies?.commentRepliesRenderer?.contents?.[0]
									?.continuationItemRenderer?.continuationEndpoint?.continuationCommand?.token;
							return;
						}
					}
				);
			}

			comments.push({
				comment: e.content.content,
				comment_id: e.commentId,
				continuationCommand: contCommand || null
			});
		});

		return {
			response,
			comment_info: comments,
			nextContinuationCommand: continuationCommand
		};
	} catch (error) {
		console.log(`Error -> getComments`);
		return null;
	}
};

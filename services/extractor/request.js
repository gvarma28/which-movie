import axios from "axios";

export const initialRequest = async (url) => {
  try {
    let config = {
      method: "GET",
      maxBodyLength: Infinity,
      url,
    };
    let response = await axios.request(config);
    if (!response || response.status !== 200) {
      throw new Error(`Invalid Response Received`, { status: response.status });
    }
    response = response.data.replaceAll("\n", "");
    response = response.split("var ytInitialData = ")[1];
    response = response.split(";</script>")[0];
    response = JSON.parse(response);
    return response;
  } catch (error) {
    console.log("Error -> initialRequest");
    return null;
  }
};

export const getComments = async (continuationCommand) => {
  try {
    let data = JSON.stringify({
      context: {
        client: {
          clientName: "WEB",
          clientVersion: "2.20240731.04.00",
        },
      },
      continuation: continuationCommand,
    });
    const config = {
      method: "post",
      maxBodyLength: Infinity,
      url: "https://www.youtube.com/youtubei/v1/browse?prettyPrint=false",
      data: data,
      headers: {
        "Content-Type": "application/json",
      },
    };
    const response = await axios.request(config);
    if (!response || response.status !== 200) {
      throw new Error(`Invalid Response Received`, { status: response.status });
    }

    const continuations =
      response?.data?.onResponseReceivedEndpoints[1]
        ?.reloadContinuationItemsCommand?.continuationItems ||
      response?.data?.onResponseReceivedEndpoints[0]
        ?.appendContinuationItemsAction?.continuationItems ||
      null;

    continuationCommand =
      continuations?.[20]?.continuationItemRenderer?.continuationEndpoint
        ?.continuationCommand?.token || null;
    // filter to have only comments elements
    const commentEntityPayload =
      response.data.frameworkUpdates.entityBatchUpdate.mutations.filter(
        (e) => e.payload.commentEntityPayload
      );
    // further map to have only comments elements
    const commentsObject = commentEntityPayload.map(
      (e) => e.payload.commentEntityPayload.properties
    );

    const comments = [];
    commentsObject.forEach((e) => {
      let contCommand;
      if (response?.data?.onResponseReceivedEndpoints?.[1]) {
        response?.data?.onResponseReceivedEndpoints?.[1]?.reloadContinuationItemsCommand?.continuationItems.forEach(
          (j) => {
            if (
              j?.commentThreadRenderer?.replies?.commentRepliesRenderer?.targetId?.includes(
                e.commentId
              )
            ) {
              contCommand =
                j?.commentThreadRenderer?.replies?.commentRepliesRenderer
                  ?.contents?.[0]?.continuationItemRenderer
                  ?.continuationEndpoint?.continuationCommand?.token;
              return;
            }
          }
        );
      } else {
        response?.data?.onResponseReceivedEndpoints?.[0]?.appendContinuationItemsAction?.continuationItems.forEach(
          (j) => {
            if (
              j?.commentThreadRenderer?.replies?.commentRepliesRenderer?.targetId?.includes(
                e.commentId
              )
            ) {
              contCommand =
                j?.commentThreadRenderer?.replies?.commentRepliesRenderer
                  ?.contents?.[0]?.continuationItemRenderer
                  ?.continuationEndpoint?.continuationCommand?.token;
              return;
            }
          }
        );
      }

      comments.push({
        comment: e.content.content,
        comment_id: e.commentId,
        continuationCommand: contCommand,
      });
    });

    return {
      response,
      comment_info: comments,
      nextContinuationCommand: continuationCommand,
    };
  } catch (error) {
    console.log(`Error -> getComments`);
    return null;
  }
};

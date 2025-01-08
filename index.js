import axios from "axios";

const initialRequest = async (url) => {
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
};

const getComments = async (continuationCommand) => {
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

  const comments = commentsObject.mcap((e) => {
    return { comment: e.content.content, comment_id: e.commentId };
  });

  return {
    response,
    comment_info: comments,
    continuationCommand,
  };
};

const main = async () => {
  const response = await initialRequest(
    "https://www.youtube.com/shorts/-BYMUKycQq8"
  );

  let continuationCommand =
    response.engagementPanels[0].engagementPanelSectionListRenderer.header
      .engagementPanelTitleHeaderRenderer.menu.sortFilterSubMenuRenderer
      .subMenuItems[0].serviceEndpoint.continuationCommand.token;

  let continuations;

  let res;
  let iteration = 0;
  while (true) {
    iteration++;
    if (iteration === 10 || !continuationCommand) break;

    const { comment_info, nextContinuationCommand } = await getComments(
      continuationCommand
    );
    continuationCommand = nextContinuationCommand;

    res = comment_info.filter((e) => e.comment.includes("name"));
    if (res && res.length > 0) break;
  }

  console.log(res);
};

main();

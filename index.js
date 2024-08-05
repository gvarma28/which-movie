import axios from "axios";

let config = {
  method: "get",
  maxBodyLength: Infinity,
  url: "https://www.youtube.com/shorts/-BYMUKycQq8",
};

let response = await axios.request(config);
if (!response || response.status !== 200) {
  throw new Error(`Invalid Response Received`, { status: response.status });
}
response = response.data.replaceAll("\n", "");
response = response.split("var ytInitialData = ")[1];
response = response.split(";</script>")[0];
response = JSON.parse(response);

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
  let data = JSON.stringify({
    context: {
      client: {
        clientName: "WEB",
        clientVersion: "2.20240731.04.00",
      },
    },
    continuation: continuationCommand,
  });
  config = {
    method: "post",
    maxBodyLength: Infinity,
    url: "https://www.youtube.com/youtubei/v1/browse?prettyPrint=false",
    data: data,
    headers: {
      "Content-Type": "application/json",
    },
  };

  response = await axios.request(config);
  if (!response || response.status !== 200) {
    throw new Error(`Invalid Response Received`, { status: response.status });
  }

  continuations =
    response?.data?.onResponseReceivedEndpoints[1]
      ?.reloadContinuationItemsCommand?.continuationItems ||
    response?.data?.onResponseReceivedEndpoints[0]
      ?.appendContinuationItemsAction ?.continuationItems ||
    null;

  continuationCommand =
    continuations?.[20]?.continuationItemRenderer?.continuationEndpoint
      ?.continuationCommand?.token || null;
  // filter to have only comments elements
  response = response.data.frameworkUpdates.entityBatchUpdate.mutations.filter(
    (e) => e.payload.commentEntityPayload
  );
  // further map to have only comments elements
  let commentsObject = response.map(
    (e) => e.payload.commentEntityPayload.properties
  );

  let comments = commentsObject.map((e) => e.content.content);


  res = commentsObject.filter((e) => e.content.content.includes("name"));
  if (res && res.length > 0) break;
}

console.log(res);

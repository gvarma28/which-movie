import { initialRequest, getComments } from "./request.js";

/**
 * @returns all the relevant comments from the yt shorts
 */
const extractor = async () => {
  const response = await initialRequest(
    "https://www.youtube.com/shorts/6Rrb0GohNoY"
    // "https://www.youtube.com/shorts/16tWbpk8sws"
  );
  if (!response) return null;

  // CCStack stands for continuationCommand stack
  const CCStack = [];
  CCStack.push(
    response.engagementPanels[0].engagementPanelSectionListRenderer.header
      .engagementPanelTitleHeaderRenderer.menu.sortFilterSubMenuRenderer
      .subMenuItems[0].serviceEndpoint.continuationCommand.token
  );

  let comments = [];
  let iteration = 0;

  while (true) {
    iteration++;
    if (iteration === 30 || !CCStack.length) break;

    const { comment_info, nextContinuationCommand } = await getComments(
      CCStack.pop()
    );
    if (!comment_info && !nextContinuationCommand) {
      console.log("Got an invalid getComments response: breaking");
      break;
    }

    if (nextContinuationCommand) CCStack.push(nextContinuationCommand);

    const movieMentionRegex =
      /(?:\b(movie|film|cinema|show|series|watched|saw|seen|about|called|name)\s+|(?:"([^"]+)"|'([^']+)'))/gi;
    const movieMention = comment_info.filter((e) =>
      movieMentionRegex.test(e.comment)
    );

    const askingForMovieRegex =
      /\b(what.s|which|can|anybody|please)\s+(movie|show|series|scene|film|is|was|this|that|it|tell)\??|name\s+(of|this|that)\s+(movie|show|series)\??/gi;
    const askingForMovie = comment_info.filter((e) =>
      askingForMovieRegex.test(e.comment)
    );

    comments.push(...movieMention, ...askingForMovie);

    // filter for comment level continuationCommand to go into the comment: to explore 'what is this movie scenarios'
    (askingForMovie?.filter((e) => e.continuationCommand) || [])?.forEach(
      (e) => {
        CCStack.push(e?.continuationCommand);
      }
    );
  }

  console.log(comments);
};

extractor();

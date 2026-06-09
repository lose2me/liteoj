import { http } from "./http";

export async function joinProblemset(problemsetID: number, password?: string) {
  return http.post(
    `/problemsets/${problemsetID}/join`,
    password ? { password } : {},
  );
}

export class Solution {
  solve(inputData: unknown): string {
    const { name } = inputData as { name: string };
    return `Hello ${name}`;
  }
}

/**
 * lib/hooks/useApiMutation.ts
 *
 * Thin wrapper over TanStack Query's useMutation, specifically for the
 * "do a write, then invalidate whatever queries are now stale" pattern
 * that covers almost every mutation in this app (create post, follow,
 * join group, mark notification read, ...).
 *
 * Why this exists: without it, every mutation site repeats
 *   const queryClient = useQueryClient();
 *   useMutation({ mutationFn, onSuccess: () => { queryClient.invalidateQueries(...) } })
 * with slightly different invalidation lists. This hook makes the
 * invalidation list the ONLY thing that varies per call site.
 *
 * Usage:
 *   const createPostMutation = useApiMutation({
 *     mutationFn: createPost,
 *     invalidateKeys: [queryKeys.feed()],
 *   });
 *
 *   createPostMutation.mutate(formData);
 *
 * For mutations where invalidation depends on the mutation's variables
 * (e.g. deleteComment needs to know which topicId to invalidate),
 * pass a function instead of a static array:
 *
 *   const deleteCommentMutation = useApiMutation({
 *     mutationFn: (vars: { commentId: number }) => deleteComment(vars.commentId),
 *     invalidateKeys: (vars) => [queryKeys.comments(vars.topicId)],
 *   });
 */

import { useMutation, useQueryClient, type UseMutationOptions } from '@tanstack/react-query';

type InvalidateKeys<TVariables> =
  | readonly (readonly unknown[])[]
  | ((variables: TVariables, data: unknown) => readonly (readonly unknown[])[]);

interface UseApiMutationOptions<TData, TVariables> extends Omit<
  UseMutationOptions<TData, Error, TVariables>,
  'onSuccess'
> {
  invalidateKeys?: InvalidateKeys<TVariables>;
  /** Runs after invalidation, e.g. for a router.push or toast */
  onSuccess?: (data: TData, variables: TVariables) => void;
}

export function useApiMutation<TData, TVariables>({
  invalidateKeys,
  onSuccess,
  ...mutationOptions
}: UseApiMutationOptions<TData, TVariables>) {
  const queryClient = useQueryClient();

  return useMutation<TData, Error, TVariables>({
    ...mutationOptions,
    onSuccess: async (data, variables) => {
      if (invalidateKeys) {
        const keys =
          typeof invalidateKeys === 'function' ? invalidateKeys(variables, data) : invalidateKeys;

        // invalidateQueries returns a Promise that resolves once the
        // resulting refetch(es) settle. Awaiting Promise.all here (instead
        // of firing-and-forgetting each call in a forEach) means:
        //   1. All invalidated keys refetch concurrently, not sequentially.
        //   2. mutateAsync() callers (e.g. CreatePostForm awaiting the
        //      mutation before router.push) can be sure fresh data is
        //      already in cache by the time onSuccess-driven navigation
        //      happens — no flash of stale content on the page you land on.
        //   3. isPending on the mutation stays true until invalidation is
        //      fully done, so a submit button's disabled state covers the
        //      whole operation, not just the initial request.
        await Promise.all(keys.map((key) => queryClient.invalidateQueries({ queryKey: key })));
      }
      onSuccess?.(data, variables);
    },
  });
}

import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { getConfig, saveConfig } from '../../../../modules/setting';
import type { Config } from '../../../../modules/setting/type';

export function useSettings() {
  const queryClient = useQueryClient();

  const configQuery = useQuery({
    queryKey: ['config'],
    queryFn: getConfig,
  });

  const saveMutation = useMutation({
    mutationFn: (config: Config) => saveConfig(config),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['config'] });
    },
  });

  return {
    config: configQuery.data,
    isLoading: configQuery.isLoading,
    isSaving: saveMutation.isPending,
    save: saveMutation.mutateAsync,
  };
}

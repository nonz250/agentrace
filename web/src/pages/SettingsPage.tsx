import { useQuery } from '@tanstack/react-query'
import { ApiKeyList } from '@/components/settings/ApiKeyList'
import { ApiKeyForm } from '@/components/settings/ApiKeyForm'
import { ProfileForm } from '@/components/settings/ProfileForm'
import { Spinner } from '@/components/ui/Spinner'
import * as keysApi from '@/api/keys'

export function SettingsPage() {
  const { data, isLoading, error } = useQuery({
    queryKey: ['keys'],
    queryFn: keysApi.getKeys,
  })

  return (
    <div>
      <h1 className="mb-6 text-2xl font-semibold text-gray-900">Settings</h1>

      <div className="space-y-8">
        <section>
          <h2 className="mb-4 text-lg font-medium text-gray-900">Profile</h2>
          <div className="rounded-xl border border-gray-200 bg-white p-6">
            <ProfileForm />
          </div>
        </section>

        <section>
          <h2 className="mb-4 text-lg font-medium text-gray-900">API Keys</h2>

          {isLoading ? (
            <div className="flex justify-center py-8">
              <Spinner />
            </div>
          ) : error ? (
            <div className="rounded-xl border border-red-200 bg-red-50 p-4 text-red-700">
              Failed to load API keys: {error.message}
            </div>
          ) : (
            <ApiKeyList keys={data?.keys || []} />
          )}
        </section>

        <section>
          <h2 className="mb-4 text-lg font-medium text-gray-900">
            Create New API Key
          </h2>
          <div className="rounded-xl border border-gray-200 bg-white p-6">
            <ApiKeyForm />
          </div>
        </section>
      </div>
    </div>
  )
}

require 'json'
require 'spec_helper'
require 'cli/client/errands_client'

describe Bosh::Cli::Client::ErrandsClient do
  subject(:client) { described_class.new(director) }
  let(:director) { instance_double('Bosh::Cli::Client::Director') }

  describe '#run_errand' do
    it 'tells director to run errand that is part of a deployment' do
      allow(director).to receive(:get_task_result_log).and_return('{}')

      expect(director).to receive(:request_and_track)
        .with(
          :post,
          '/deployments/fake-deployment-name/errands/fake-errand-name/runs',
          { content_type: 'application/json', payload: '{}' },
        )
        .and_return([:done, 'fake-task-id'])

      client.run_errand('fake-deployment-name', 'fake-errand-name')
    end

    context 'when task status is :done' do
      before { allow(director).to receive(:request_and_track).and_return([:done, 'fake-task-id']) }

      it 'fetches the output for the task and return an errand result' do
        raw_task_output = JSON.dump(exit_code: 123, stdout: 'fake-stdout', stderr: 'fake-stderr')

        expect(director).to receive(:get_task_result_log).
          with('fake-task-id').
          and_return("#{raw_task_output}\n")

        status, task_id, actual_result = client.run_errand('fake-deployment-name', 'fake-errand-name')
        expect(status).to eq(:done)
        expect(task_id).to eq('fake-task-id')
        expect(actual_result).to eq(described_class::ErrandResult.new(123, 'fake-stdout', 'fake-stderr'))
      end
    end

    context 'when task status is not :done (e.g. error, etc)' do
      before { allow(director).to receive(:request_and_track).and_return([:not_done, 'fake-task-id']) }

      it 'returns status, task_id and result' do
        status, task_id, result = client.run_errand('fake-deployment-name', 'fake-errand-name')
        expect(status).to eq(:not_done)
        expect(task_id).to eq('fake-task-id')
        expect(result).to be(nil)
      end
    end
  end
end

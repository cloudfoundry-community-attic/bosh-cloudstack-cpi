require 'spec_helper'

module Bosh::Director
  describe Jobs::RunErrand do
    subject { described_class.new('fake-dep-name', 'fake-errand-name') }

    before do
      App.stub_chain(:instance, :blobstores, :blobstore).
        with(no_args).
        and_return(blobstore)
    end
    let(:blobstore) { instance_double('Bosh::Blobstore::Client') }

    describe 'Resque job class expectations' do
      let(:job_type) { :run_errand }
      it_behaves_like 'a Resque job'
    end

    describe '#perform' do
      context 'when deployment exists' do
        let!(:deployment_model) do
          Models::Deployment.make(
            name: 'fake-dep-name',
            manifest: "---\nmanifest: true",
          )
        end

        before { allow(Config).to receive(:event_log).with(no_args).and_return(event_log) }
        let(:event_log) { instance_double('Bosh::Director::EventLog::Log') }

        before do
          allow(DeploymentPlan::Planner).to receive(:parse).
            with({'manifest' => true}, event_log, {}).
            and_return(deployment)
        end
        let(:deployment) { instance_double('Bosh::Director::DeploymentPlan::Planner') }

        context 'when job representing an errand exists' do
          before { allow(deployment).to receive(:job).with('fake-errand-name').and_return(job) }
          let(:job) { instance_double('Bosh::Director::DeploymentPlan::Job', name: 'fake-errand-name') }

          context "when job can run as an errand (usually means lifecycle: errand)" do
            before { allow(job).to receive(:can_run_as_errand?).and_return(true) }

            context 'when job has at least 1 instance' do
              before { allow(job).to receive(:instances).with(no_args).and_return([instance]) }
              let(:instance) { instance_double('Bosh::Director::DeploymentPlan::Instance') }

              before { allow(Config).to receive(:result).with(no_args).and_return(result_file) }
              let(:result_file) { instance_double('Bosh::Director::TaskResultFile') }

              before { allow(job).to receive(:resource_pool).with(no_args).and_return(resource_pool) }
              let(:resource_pool) { instance_double('Bosh::Director::DeploymentPlan::ResourcePool') }

              it 'runs an errand and returns short result description' do
                deployment_preparer = instance_double('Bosh::Director::Errand::DeploymentPreparer')
                expect(Errand::DeploymentPreparer).to receive(:new).
                  with(deployment, job, event_log, subject).
                  and_return(deployment_preparer)

                expect(deployment_preparer).to receive(:prepare_deployment).with(no_args).ordered
                expect(deployment_preparer).to receive(:prepare_job).with(no_args).ordered

                rp_updater = instance_double('Bosh::Director::ResourcePoolUpdater')
                expect(ResourcePoolUpdater).to receive(:new).
                  with(resource_pool).
                  and_return(rp_updater)

                rp_manager = instance_double('Bosh::Director::DeploymentPlan::ResourcePools')
                expect(DeploymentPlan::ResourcePools).to receive(:new).
                  with(event_log, [rp_updater]).
                  and_return(rp_manager)

                expect(rp_manager).to receive(:update).with(no_args).ordered

                job_manager = instance_double('Bosh::Director::Errand::JobManager')
                expect(Errand::JobManager).to receive(:new).
                  with(deployment, job, blobstore, event_log).
                  and_return(job_manager)

                expect(job_manager).to receive(:update_instances).with(no_args).ordered

                runner = instance_double('Bosh::Director::Errand::Runner')
                expect(Errand::Runner).to receive(:new).
                  with(job, result_file, be_a(Api::InstanceManager), event_log).
                  and_return(runner)

                expect(runner).to receive(:run).
                  with(no_args).
                  ordered.
                  and_return('fake-result-short-description')

                expect(job_manager).to receive(:delete_instances).with(no_args).ordered
                expect(rp_manager).to receive(:refill).with(no_args).ordered

                expect(subject.perform).to eq('fake-result-short-description')
              end
            end

            context 'when job representing an errand has 0 instances' do
              before { allow(job).to receive(:instances).with(no_args).and_return([]) }

              it 'raises an error because errand cannot be run on a job without 0 instances' do
                expect {
                  subject.perform
                }.to raise_error(InstanceNotFound, %r{fake-errand-name/0.*doesn't exist})
              end
            end
          end

          context "when job cannot run as an errand (e.g. marked as 'lifecycle: service')" do
            before { allow(job).to receive(:can_run_as_errand?).and_return(false) }

            it 'raises an error because non-errand jobs cannot be used with run errand cmd' do
              expect {
                subject.perform
              }.to raise_error(RunErrandError, /Job `fake-errand-name' is not an errand/)
            end
          end
        end

        context 'when job representing an errand does not exist' do
          before { allow(deployment).to receive(:job).with('fake-errand-name').and_return(nil) }

          it 'raises an error because user asked to run an unknown errand' do
            expect {
              subject.perform
            }.to raise_error(JobNotFound, %r{fake-errand-name.*doesn't exist})
          end
        end
      end

      context 'when deployment does not exist' do
        it 'raises an error' do
          expect {
            subject.perform
          }.to raise_error(DeploymentNotFound, %r{fake-dep-name.*doesn't exist})
        end
      end
    end
  end
end

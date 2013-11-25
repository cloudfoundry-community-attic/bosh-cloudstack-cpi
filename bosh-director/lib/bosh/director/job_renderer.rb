require 'bosh/director/job_template_loader'
require 'bosh/director/job_instance_renderer'
require 'bosh/director/rendered_job_instance_hasher'
require 'bosh/director/rendered_templates_uploader'

module Bosh::Director
  class JobRenderer
    # @param [DeploymentPlan::Job]
    def initialize(job)
      @job = job
      job_template_loader = JobTemplateLoader.new
      @instance_renderer = JobInstanceRenderer.new(@job, job_template_loader)
    end

    def render_job_instances
      @job.instances.each do |instance|
        rendered_templates = @instance_renderer.render(instance)

        uploader = RenderedTemplatesUploader.new
        uploader.upload(rendered_templates)

        hasher = RenderedJobInstanceHasher.new(rendered_templates)
        instance.configuration_hash = hasher.configuration_hash
        instance.template_hashes = hasher.template_hashes
      end
    end
  end
end

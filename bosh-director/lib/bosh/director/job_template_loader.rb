require 'bosh/director/job_template_renderer'

module Bosh::Director
  SrcFileTemplate = Struct.new(:src_name, :dest_name, :erb_file)

  class JobTemplateLoader
    def process(job_template)
      template_dir = extract_template(job_template)
      manifest = Psych.load_file(File.join(template_dir, 'job.MF'))

      monit_template = erb(File.join(template_dir, 'monit'))
      monit_template.filename = File.join(job_template.name, 'monit')

      templates = []

      manifest.fetch('templates', {}).each_pair do |src_name, dest_name|
        erb_file = erb(File.join(template_dir, 'templates', src_name))
        erb_file.filename = File.join(job_template.name, src_name)
        templates << SrcFileTemplate.new(src_name, dest_name, erb_file)
      end

      JobTemplateRenderer.new(job_template.name, monit_template, templates)
    ensure
      FileUtils.rm_rf(template_dir) if template_dir
    end

    private

    def extract_template(job_template)
      temp_path = job_template.download_blob
      template_dir = Dir.mktmpdir('template_dir')

      output = `tar -C #{template_dir} -xzf #{temp_path} 2>&1`
      if $?.exitstatus != 0
        raise JobTemplateUnpackFailed,
              "Cannot unpack `#{job_template.name}' job template, " +
                "tar returned #{$?.exitstatus}, " +
                "tar output: #{output}"
      end

      template_dir
    ensure
      FileUtils.rm_f(temp_path) if temp_path
    end

    def erb(path)
      ERB.new(File.read(path))
    end
  end
end

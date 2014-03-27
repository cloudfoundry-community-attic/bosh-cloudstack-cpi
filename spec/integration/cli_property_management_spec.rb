require 'spec_helper'

describe 'cli: property management', type: :integration do
  with_reset_sandbox_before_each

  it 'can get/set/unset deployment properties' do
    manifest = Bosh::Spec::Deployments.simple_manifest
    manifest['jobs'] = []
    deploy_simple(manifest_hash: manifest)

    expect(run_bosh('set property foo bar')).to match /Property `foo' set to `bar'/
    expect(run_bosh('get property foo')).to match /Property `foo' value is `bar'/
    expect(run_bosh('set property foo baz')).to match /Property `foo' set to `baz'/
    expect(run_bosh('unset property foo')).to match /Property `foo' has been unset/
    expect(run_bosh('get property foo', failure_expected: true)).to match /Error 110003: Property `foo' not found/
    expect(run_bosh('unset property foo', failure_expected: true)).to match /Error 110003: Property `foo' not found/

    run_bosh('set property nats.user admin')
    run_bosh('set property nats.password pass')

    props = run_bosh('properties --terse')
    expect(props).to match /nats.user\tadmin/
    expect(props).to match /nats.password\tpass/
  end
end

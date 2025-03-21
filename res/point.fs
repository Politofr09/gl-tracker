#version 330

in vec2 fragTexCoord;
out vec4 fragColor;

uniform sampler2D texture0;
uniform float time;

void main()
{
    fragColor = texture(texture0, fragTexCoord);
}

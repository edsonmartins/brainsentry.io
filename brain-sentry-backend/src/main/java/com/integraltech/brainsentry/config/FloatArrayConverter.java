package com.integraltech.brainsentry.config;

import jakarta.persistence.AttributeConverter;
import jakarta.persistence.Converter;
import java.nio.ByteBuffer;
import java.nio.ByteOrder;

/**
 * JPA converter for float[] to byte[].
 *
 * Converts float arrays to byte arrays for storage in PostgreSQL bytea columns.
 * Each float is converted to 4 bytes using IEEE 754 format.
 */
@Converter
public class FloatArrayConverter implements AttributeConverter<float[], byte[]> {

    @Override
    public byte[] convertToDatabaseColumn(float[] attribute) {
        if (attribute == null) {
            return null;
        }
        ByteBuffer buffer = ByteBuffer.allocate(attribute.length * 4);
        buffer.order(ByteOrder.LITTLE_ENDIAN);
        for (float value : attribute) {
            buffer.putFloat(value);
        }
        return buffer.array();
    }

    @Override
    public float[] convertToEntityAttribute(byte[] dbData) {
        if (dbData == null || dbData.length % 4 != 0) {
            return null;
        }
        ByteBuffer buffer = ByteBuffer.wrap(dbData);
        buffer.order(ByteOrder.LITTLE_ENDIAN);
        float[] result = new float[dbData.length / 4];
        for (int i = 0; i < result.length; i++) {
            result[i] = buffer.getFloat();
        }
        return result;
    }
}
